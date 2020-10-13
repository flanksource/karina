package platform

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/is"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/net"
	"github.com/flanksource/commons/text"
	"github.com/flanksource/karina/manifests"
	"github.com/flanksource/karina/pkg/api"
	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/client/dns"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/k8s/proxy"
	"github.com/flanksource/karina/pkg/types"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	pg "github.com/go-pg/pg/v9"
	minio "github.com/minio/minio-go/v6"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/flanksource/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Platform struct {
	Cluster types.Cluster
	types.PlatformConfig
	MasterDiscovery MasterDiscovery
	ProvisionHook   ProvisionHook
	logger.Logger
	logFields map[string]interface{}
	k8s.Client
	//TODO: verify if ctx can be removed after refactoring has left it unused
	ctx            context.Context //nolint
	kubeConfig     []byte
	KubeConfigPath string
	ca             certs.CertificateAuthority
	ingressCA      certs.CertificateAuthority
	// Terminating is true if the cluster is in a terminating state
	Terminating bool
}

func (platform *Platform) Init() error {
	if platform.Client.GetKubeConfigBytes == nil {
		platform.Client.GetKubeConfigBytes = platform.GetKubeConfigBytes
	}
	platform.Client.GetKustomizePatches = func() ([]string, error) {
		return platform.Patches, nil
	}
	platform.Client.ApplyDryRun = platform.DryRun
	platform.Client.Trace = platform.PlatformConfig.Trace
	platform.Logger = logrus.StandardLogger().WithContext(context.Background())
	platform.Client.Logger = platform.Logger

	platform.logFields = make(map[string]interface{})
	consul := NewConsulProvider(platform)
	dns := NewDNSProvider(platform.GetDNSClient())

	if platform.NSX != nil && !platform.NSX.Disabled {
		nsx, err := NewNSXProvider(platform)
		if err != nil {
			return err
		}
		platform.MasterDiscovery = nsx
		if platform.DNS.IsEnabled() && platform.DNS.UpdateHosts {
			platform.ProvisionHook = CompositeHook{Hooks: []ProvisionHook{nsx, dns}}
		} else {
			platform.ProvisionHook = nsx
		}
	} else if platform.Consul != "" && platform.DNS.IsEnabled() && platform.DNS.UpdateHosts {
		// when both consul and DNS are specified, Consul is used for master discovery
		// and DNS used for external access
		platform.MasterDiscovery = consul
		platform.ProvisionHook = CompositeHook{Hooks: []ProvisionHook{consul, dns}}
	} else if platform.Consul != "" {
		platform.MasterDiscovery = consul
		platform.ProvisionHook = consul
	} else if platform.DNS.IsEnabled() && platform.DNS.UpdateHosts {
		platform.MasterDiscovery = dns
		platform.ProvisionHook = dns
	} else {
		platform.MasterDiscovery = KindProvider{}
		platform.ProvisionHook = CompositeHook{}
	}

	if platform.KubeConfigPath == "" {
		if os.Getenv("KUBECONFIG") == "" {
			platform.KubeConfigPath = os.ExpandEnv("$HOME/.kube/config")
		} else {
			platform.KubeConfigPath = os.Getenv("KUBECONFIG")
		}
	}
	return nil
}

func (platform *Platform) clone() *Platform {
	logFields := make(map[string]interface{})
	for k, v := range platform.logFields {
		logFields[k] = v
	}
	return &Platform{
		Cluster:         platform.Cluster,
		PlatformConfig:  platform.PlatformConfig,
		Logger:          platform.Logger,
		logFields:       logFields,
		Client:          platform.Client,
		ctx:             context.TODO(),
		MasterDiscovery: platform.MasterDiscovery,
		ProvisionHook:   platform.ProvisionHook,
		kubeConfig:      platform.kubeConfig,
		ca:              platform.ca,
		ingressCA:       platform.ingressCA,
		KubeConfigPath:  platform.KubeConfigPath,
		Terminating:     platform.Terminating,
	}
}

func (platform *Platform) WithField(key string, value interface{}) *Platform {
	copy := platform.clone()
	logger := copy.Logger.(*logrus.Entry)
	copy.logFields[key] = value
	copy.Logger = logger.WithField(key, value)
	copy.Client.Logger = copy.Logger
	return copy
}

func (platform *Platform) WithLogOutput(output io.Writer) *Platform {
	copy := platform.clone()
	logger := logrus.New()
	logger.SetOutput(output)
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	newLogger := logger.WithContext(context.Background())
	for k, v := range copy.logFields {
		newLogger = newLogger.WithField(k, v)
	}
	copy.Logger = newLogger
	copy.Client.Logger = copy.Logger
	return copy
}

func (platform *Platform) ResetMasterConnection() {
	platform.kubeConfig = nil
	platform.Client.ResetConnection()
}

// GetAPIEndpoint returns an endpoint for reaching a master node that is reachable on 6443 or
// an error otherwise
func (platform *Platform) GetAPIEndpoint() (string, error) {
	if platform.DNS.IsEnabled() {
		ip := fmt.Sprintf("k8s-api.%s", platform.Domain)
		if net.Ping(ip, 6443, 10) {
			return ip, nil
		}
		platform.Warnf("DNS endpoint is not healthy, failing back to master IP")
	}

	masters, err := platform.MasterDiscovery.GetExternalEndpoints(platform)
	if err != nil {
		return "", errors.WithMessage(err, "failed to discover any external endpoints")
	}

	if len(masters) == 0 {
		return "", fmt.Errorf("could not find any master ips")
	}

	for _, master := range masters {
		if net.Ping(master, 6443, 10) {
			return master, nil
		}
	}
	return "", fmt.Errorf("none of the masters are up: %v", masters)
}

func (platform *Platform) GetKubeConfigBytes() ([]byte, error) {
	if platform.kubeConfig != nil {
		return platform.kubeConfig, nil
	}

	context := k8s.GetCurrentClusterNameFrom(platform.KubeConfigPath)
	platform.Tracef("Current KUBECONFIG: path=%s, context=%s", platform.KubeConfigPath, context)
	if platform.Name == k8s.GetCurrentClusterNameFrom(platform.KubeConfigPath) {
		return ioutil.ReadFile(platform.KubeConfigPath)
	}

	ip, err := platform.GetAPIEndpoint()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to discover any healthy external endpoints")
	}

	platform.Debugf("Generating a new kubeconfig for %s", ip)
	kubeConfig, err := k8s.CreateKubeConfig(platform.Name, platform.GetCA(), ip, "system:masters", "admin", 24*7*time.Hour)
	if err != nil {
		return nil, err
	}
	platform.kubeConfig = kubeConfig
	return platform.kubeConfig, nil
}

// GetCA retrieves the cert.CertificateAuthority
// for the given platform, initialising it (platform.ca) if it hasn't been read from
// the specified config (platform.CA) yet.
func (platform *Platform) GetCA() certs.CertificateAuthority {
	if platform.ca != nil {
		return platform.ca
	}
	ca, err := ca.ReadCA(platform.CA)
	if err != nil {
		platform.Fatalf("Unable to open %s: %v", platform.CA.PrivateKey, err)
	}
	platform.ca = ca
	return ca
}

func (platform *Platform) ReadIngressCACertString() string {
	cert := files.SafeRead(platform.IngressCA.Cert)
	return cert
}

func (platform *Platform) GetIngressCA() certs.CertificateAuthority {
	if platform.ingressCA != nil {
		return platform.ingressCA
	}

	if platform.IngressCA == nil {
		platform.Infof("Creating self-signed CA for ingress")
		ca := certs.NewCertificateBuilder("ingress-ca").CA().Certificate
		platform.ingressCA, _ = ca.SignCertificate(ca, 1)
		return platform.ingressCA
	}
	platform.Debugf("[IngressCA] loading from disk: %s", platform.IngressCA.Cert)
	ca, err := ca.ReadCA(platform.IngressCA)
	if err != nil {
		platform.Fatalf("Unable to open Ingress CA: %v", err)
	}
	platform.Debugf("[IngressCA] read CA %s", ca.X509.Subject)
	platform.ingressCA = ca
	return ca
}

// WaitFor at least 1 master IP to be reachable
func (platform *Platform) WaitFor() error {
	if platform.DryRun {
		return nil
	}
	for {
		api, _ := platform.GetAPIEndpoint()
		if api != "" && platform.PingMaster() {
			return nil
		}
		platform.ResetMasterConnection()
		time.Sleep(5 * time.Second)
	}
}

func (platform *Platform) GetDNSClient() dns.Client {
	if !platform.DNS.IsEnabled() {
		return &dns.DummyDNSClient{
			Logger: platform.Logger,
			Zone:   "nip.io",
		}
	}

	if platform.DNS.Type == "route53" {
		dns := &dns.Route53Client{
			Logger:       platform.Logger,
			HostedZoneID: platform.DNS.Zone,
			AccessKey:    platform.DNS.AccessKey,
			SecretKey:    platform.DNS.SecretKey,
		}
		dns.Init()
		return dns
	}

	return &dns.DynamicDNSClient{
		Logger:     platform.Logger,
		Zone:       platform.DNS.Zone,
		KeyName:    platform.DNS.KeyName,
		Nameserver: platform.DNS.Nameserver,
		Key:        platform.DNS.Key,
		Algorithm:  platform.DNS.Algorithm,
	}
}

func (platform *Platform) Clone(vm types.VM, config *konfigadm.Config) (types.Machine, error) {
	if platform.DryRun {
		return types.NullMachine{}, nil
	}
	vm.Konfigadm = config

	if err := platform.ProvisionHook.BeforeProvision(platform, &vm); err != nil {
		return types.NullMachine{}, err
	}

	for _, cmd := range vm.Commands {
		config.AddCommand(cmd)
	}

	VM, err := platform.Cluster.Clone(vm, config)
	if err != nil {
		return nil, err
	}

	if err := VM.SetAttributes(map[string]string{
		"Template":    vm.Template,
		"CreatedDate": time.Now().Format("02Jan06-15:04:05"),
	}); err != nil {
		platform.Warnf("Failed to set attributes for %s: %v", vm.Name, err)
	}
	platform.Debugf("[%s] Waiting for IP", vm.Name)
	ip, err := VM.WaitForIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP for %s: %v", vm.Name, err)
	}
	vm.IP = ip
	platform.Tracef("[%s] found ip %s", vm.Name, ip)

	if err := platform.ProvisionHook.AfterProvision(platform, VM); err != nil {
		return types.NullMachine{}, err
	}

	return VM, nil
}

func (platform *Platform) DeleteNode(name string) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}
	node, err := client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	err = client.CoreV1().Nodes().Delete(node.Name, &metav1.DeleteOptions{})
	if err == nil {
		platform.Infof("[%s] deleted node", node.Name)
	}
	return err
}

func (platform *Platform) GetConsulClient() api.Consul {
	return api.Consul{
		Logger:  platform.Logger,
		Host:    fmt.Sprintf("http://%s:8500", platform.Consul),
		Service: platform.Name,
	}
}

func (platform *Platform) GetNodeNames() map[string]bool {
	client, err := platform.GetClientset()
	if err != nil {
		return nil
	}
	existingNodes := map[string]bool{}
	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil
	}
	for _, node := range nodeList.Items {
		existingNodes[node.Name] = true
	}
	return existingNodes
}

// GetKubeConfig gets the path to the admin kubeconfig, creating it if necessary
func (platform *Platform) GetKubeConfig() (string, error) {
	// if the current kubeconfig context already has a reference to the cluster
	// then we can just reuse it
	if platform.Name == k8s.GetCurrentClusterNameFrom(platform.KubeConfigPath) {
		return platform.KubeConfigPath, nil
	}

	if !is.File(platform.KubeConfigPath) {
		data, err := platform.GetKubeConfigBytes()
		if err != nil {
			return "", err
		}
		if err := ioutil.WriteFile(platform.KubeConfigPath, data, 0644); err != nil {
			return "", err
		}
	}
	return platform.KubeConfigPath, nil
}

func (platform *Platform) GetBinaryWithKubeConfig(binary string) deps.BinaryFunc {
	kubeconfig, err := platform.GetKubeConfig()
	if err != nil {
		return func(msg string, args ...interface{}) error {
			return fmt.Errorf("cannot create kubeconfig %v", err)
		}
	}
	if platform.DryRun {
		return platform.GetBinary(binary)
	}

	return deps.BinaryWithEnv(binary, platform.Versions[binary], ".bin", map[string]string{
		"KUBECONFIG": kubeconfig,
		"PATH":       os.Getenv("PATH"),
	})
}

func (platform *Platform) GetKubectl() deps.BinaryFunc {
	kubeconfig, err := platform.GetKubeConfig()
	if err != nil {
		return func(msg string, args ...interface{}) error {
			return fmt.Errorf("cannot create kubeconfig %v", err)
		}
	}
	if platform.DryRun {
		return platform.GetBinary("kubectl")
	}

	platform.Tracef("Using KUBECONFIG=%s", kubeconfig)
	return deps.BinaryWithEnv("kubectl", platform.Kubernetes.Version, ".bin", map[string]string{
		"KUBECONFIG": kubeconfig,
		"PATH":       os.Getenv("PATH"),
	})
}

func (platform *Platform) CreateTLSSecret(namespace, subDomain, secretName string) error {
	if platform.HasSecret(namespace, secretName) {
		platform.Debugf("secret/%s/%s' for %s already exists", namespace, secretName, subDomain)
		//TODO(moshloop) check certificate expiry and renew if necessary
		return nil
	}
	platform.Infof("Creating new ingress cert %s.%s", subDomain, platform.Domain)
	cert := certs.NewCertificateBuilder(subDomain + "." + platform.Domain).Server().Certificate

	cert.X509.PublicKey = cert.PrivateKey.Public()

	// we are using cert-manager so we create a very short-lived cert
	// so that services can start (with an invalid cert), and then let
	// cert-manager "renew it"
	expiry := time.Hour * 24 * 10

	signed, err := platform.GetIngressCA().Sign(cert.X509, expiry)
	if err != nil {
		return fmt.Errorf("failed to sign cert %s: %v", cert.X509.Subject.CommonName, err)
	}

	cert = &certs.Certificate{
		X509:       signed,
		PrivateKey: cert.PrivateKey,
	}
	return platform.CreateOrUpdateSecret(secretName, namespace, cert.AsTLSSecret())
}

func (platform *Platform) CreateIngressCertificate(subDomain string) (*certs.Certificate, error) {
	platform.Infof("Creating new ingress cert %s.%s", subDomain, platform.Domain)
	cert := certs.NewCertificateBuilder(subDomain + "." + platform.Domain).Server().Certificate
	return platform.GetIngressCA().SignCertificate(cert, 3)
}

func (platform *Platform) NewSelfSigned(domain string) *certs.Certificate {
	platform.Infof("Creating new self signed cert %s", domain)
	cert := certs.NewCertificateBuilder(domain).Server().Certificate
	cert, _ = cert.SignCertificate(cert, 10)
	return cert
}

func (platform *Platform) CreateInternalCertificate(service string, namespace string, clusterDomain string) (*certs.Certificate, error) {
	domain := fmt.Sprintf("%s.%s.svc.%s", service, namespace, clusterDomain)
	platform.Infof("Creating new internal certificate %s", domain)
	cert := certs.NewCertificateBuilder(domain).Server().Certificate
	return platform.GetIngressCA().SignCertificate(cert, 5)
}

func (platform *Platform) GetResourceByName(file string, pkg string) (string, error) {
	var raw string
	var err error
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	raw, err = manifests.FSString(false, file)
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (platform *Platform) Template(file string, pkg string) (string, error) {
	raw, err := platform.GetResourceByName(file, pkg)
	if err != nil {
		return "", fmt.Errorf("could not find %s: %v", file, err)
	}
	if strings.HasSuffix(file, ".raw") {
		return raw, nil
	}
	template, err := text.Template(raw, platform.PlatformConfig)
	if err != nil {
		data, _ := yaml.Marshal(platform.PlatformConfig)
		platform.Debugf("Error templating %s: %s", file, console.StripSecrets(string(data)))
		return "", err
	}
	return template, nil
}

func (platform *Platform) GetResourcesByDir(path string, pkg string) (map[string]http.File, error) {
	out := make(map[string]http.File)
	fs := manifests.FS(false)
	dir, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("getResourcesByDir: failed to open fs: %v", err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("getResourcesByDir: failed to read dir: %v", err)
	}

	for _, info := range files {
		file, err := fs.Open(path + "/" + info.Name())
		if err != nil {
			return nil, fmt.Errorf("getResourcesByDir: failed to open fs: %v", err)
		}
		out[info.Name()] = file
	}
	return out, nil
}

func (platform *Platform) ExposeIngress(namespace, service string, port int, annotations map[string]string) error {
	return platform.Client.ExposeIngress(namespace, service, fmt.Sprintf("%s.%s", service, platform.Domain), port, annotations)
}

func (platform *Platform) ApplyCRD(namespace string, specs ...k8s.CRD) error {
	for _, spec := range specs {
		data, err := yaml.Marshal(spec)
		if err != nil {
			return fmt.Errorf("applyCRD: failed to marshal yaml specs: %v", err)
		}

		if err := platform.ApplyText(namespace, string(data)); err != nil {
			return err
		}
	}
	return nil
}

func (platform *Platform) ApplyText(namespace string, specs ...string) error {
	kustomize, err := platform.GetKustomize()
	if err != nil {
		return err
	}
	for _, spec := range specs {
		items, err := kustomize.Kustomize(namespace, []byte(spec))
		if err != nil {
			return err
		}
		if err := platform.Apply(namespace, items...); err != nil {
			return err
		}
	}
	return nil
}

func (platform *Platform) WaitForNamespace(ns string, timeout time.Duration) {
	client, err := platform.GetClientset()
	if err != nil {
		return
	}
	k8s.WaitForNamespace(client, ns, timeout)
}

func (platform *Platform) DeleteSpecs(namespace string, specs ...string) error {
	if platform.TerminationProtection {
		platform.Debugf("Skipping deletion of resources when termination protection is enabled ")
		return nil
	}
	for _, spec := range specs {
		template, err := platform.Template(spec, "manifests")
		if err != nil {
			return err
		}
		objects, err := k8s.GetUnstructuredObjects([]byte(template))
		if err != nil {
			return err
		}
		// reverse the order of the objects so that they can be deleted in reverse-order
		for i, j := 0, len(objects)-1; i < j; i, j = i+1, j-1 {
			objects[i], objects[j] = objects[j], objects[i]
		}

		platform.Debugf("Deleting %s", console.Redf("%s", spec))
		for _, object := range objects {
			if err := platform.Get(object.GetNamespace(), object.GetName(), &object); err != nil {
				platform.Tracef("resources already deleted skipping, %v", err)
				return nil
			}

			platform.Tracef("Deleting %s/%s/%s", object.GetNamespace(), object.GetKind(), object.GetName())
			if err := platform.DeleteUnstructured(namespace, &object); err != nil {
				return err
			}
		}
	}
	return nil
}

func (platform *Platform) ApplySpecs(namespace string, specs ...string) error {
	for _, spec := range specs {
		platform.Debugf("Applying %s", console.Greenf("%s", spec))
		template, err := platform.Template(spec, "manifests")
		if err != nil {
			return fmt.Errorf("applySpecs: failed to template manifests: %v", err)
		}

		if err := platform.ApplyText(namespace, template); err != nil {
			return err
		}
	}
	return nil
}

func (platform *Platform) GetBinaryWithEnv(name string, env map[string]string) deps.BinaryFunc {
	if platform.DryRun {
		return func(msg string, args ...interface{}) error {
			platform.Tracef("CMD: "+fmt.Sprintf("%s", env)+" .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	return deps.BinaryWithEnv(name, platform.Versions[name], ".bin", env)
}

func (platform *Platform) GetBinary(name string) deps.BinaryFunc {
	if platform.DryRun {
		return func(msg string, args ...interface{}) error {
			platform.Tracef("CMD: .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	os.Mkdir(".bin", 0755) //nolint: errcheck
	return deps.Binary(name, platform.Versions[name], ".bin")
}

func (platform *Platform) GetOrCreateBucket(name string) error {
	if platform.ApplyDryRun {
		platform.Debugf("[dry-run] creating bucket %s", name)
		return nil
	}
	s3Client, err := platform.GetS3Client()
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %v", err)
	}

	exists, err := s3Client.BucketExists(name)
	if err != nil {
		return fmt.Errorf("failed to check S3 bucket: %v", err)
	}
	if !exists {
		platform.Infof("Creating s3://%s", name)
		if err := s3Client.MakeBucket(name, platform.S3.Region); err != nil {
			return fmt.Errorf("failed to create S3 bucket: %v", err)
		}
	}
	return nil
}

func (platform *Platform) GetOrCreateBucketFor(conn types.S3Connection, name string) error {
	if platform.ApplyDryRun {
		platform.Debugf("[dry-run] creating bucket %s", name)
		return nil
	}
	s3Client, err := platform.GetS3ClientFor(conn)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %v", err)
	}

	exists, err := s3Client.BucketExists(name)
	if err != nil {
		return fmt.Errorf("failed to check S3 bucket: %v", err)
	}
	if !exists {
		platform.Infof("Creating s3://%s", name)
		if err := s3Client.MakeBucket(name, platform.S3.Region); err != nil {
			return fmt.Errorf("failed to create S3 bucket: %v", err)
		}
	}
	return nil
}

func (platform *Platform) GetProxyTransport(endpoint string) (*http.Transport, error) {
	client, _ := platform.GetClientset()
	name := strings.Split(endpoint, ".")[0]
	ns := strings.Split(endpoint, ".")[1]
	svc, err := client.CoreV1().Services(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	port := int(svc.Spec.Ports[0].Port)
	dialer, _ := platform.GetProxyDialer(proxy.Proxy{
		Namespace:    ns,
		Kind:         "pods",
		ResourceName: name + "-0",
		Port:         port,
	})
	platform.Infof("proxying %s through svc=%s ns=%s port=%d", endpoint, name, ns, port)
	return &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}, nil
}

func (platform *Platform) GetS3ClientFor(conn types.S3Connection) (*minio.Client, error) {
	endpoint := conn.Endpoint
	endpoint = strings.ReplaceAll(endpoint, "http://", "")
	endpoint = strings.ReplaceAll(endpoint, "https://", "")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var err error
	if strings.HasSuffix(endpoint, ".svc") || strings.Contains(endpoint, ".svc:") {
		tr, err = platform.GetProxyTransport(endpoint)
		if err != nil {
			return nil, err
		}
	}

	if endpoint == "" {
		endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", conn.Region)
	}

	s3, err := minio.New(endpoint, conn.AccessKey, conn.SecretKey, false)
	if err != nil {
		return nil, err
	}
	s3.SetCustomTransport(tr)
	return s3, nil
}

func (platform *Platform) GetS3Client() (*minio.Client, error) {
	return platform.GetS3ClientFor(platform.S3.S3Connection)
}

func (platform *Platform) OpenDB(namespace, clusterName, databaseName string) (*pg.DB, error) {
	client, _ := platform.GetClientset()
	opts := metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster-name=%s,spilo-role=master", clusterName)}
	pods, err := client.CoreV1().Pods(namespace).List(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get master pod for cluster %s: %v", clusterName, err)
	}

	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("expected 1 pod for spilo-role=master got %d", len(pods.Items))
	}

	secretName := fmt.Sprintf("app.%s.credentials", clusterName)
	secret := platform.GetSecret("postgres-operator", secretName)
	if secret == nil {
		return nil, fmt.Errorf("%s not found", secretName)
	}

	dialer, err := platform.GetProxyDialer(proxy.Proxy{
		Namespace:    namespace,
		Kind:         "pods",
		ResourceName: pods.Items[0].Name,
		Port:         5432,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get proxy dialer")
	}

	pgdb := pg.Connect(&pg.Options{
		User:     string((*secret)["username"]),
		Password: string((*secret)["password"]),
		Dialer:   dialer.DialContext,
		Database: databaseName,
	})

	return pgdb, nil
}

func (platform *Platform) CreateOrUpdateNamespace(name string, labels map[string]string, annotations map[string]string) error {
	// set default labels
	defaultLabels := make(map[string]string)
	defaultLabels["openpolicyagent.org/webhook"] = "ignore"
	if labels != nil {
		for k, v := range defaultLabels {
			labels[k] = v
		}
	} else {
		labels = defaultLabels
	}
	// set default annotations
	defaultAnnotations := make(map[string]string)
	defaultAnnotations["com.flanksource.infra.logs/enabled"] = "true"
	if annotations != nil {
		for k, v := range defaultAnnotations {
			annotations[k] = v
		}
	} else {
		annotations = defaultAnnotations
	}

	return platform.Client.CreateOrUpdateNamespace(name, labels, annotations)
}

func (platform *Platform) CreateOrUpdateWorkloadNamespace(name string, labels map[string]string, annotations map[string]string) error {
	return platform.Client.CreateOrUpdateNamespace(name, labels, annotations)
}

func (platform *Platform) DefaultNamespaceLabels() map[string]string {
	annotations := map[string]string{
		"openpolicyagent.org/webhook": "ignore",
	}
	return annotations
}

func (platform *Platform) DefaultNamespaceAnnotations() map[string]string {
	annotations := map[string]string{
		"com.flanksource.infra.logs/enabled": "true",
	}
	return annotations
}

func (platform *Platform) IsMaster(machine types.TagInterface) bool {
	return machine.GetTags()["Role"] == platform.Name+"-masters"
}
