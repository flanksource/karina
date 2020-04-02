package platform

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	"github.com/flanksource/commons/net"
	"github.com/flanksource/commons/text"
	"github.com/flanksource/yaml"
	minio "github.com/minio/minio-go/v6"
	log "github.com/sirupsen/logrus"

	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/manifests"
	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/api/postgres"
	"github.com/moshloop/platform-cli/pkg/client/dns"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/nsx"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/templates"
)

type Platform struct {
	Cluster types.Cluster
	types.PlatformConfig

	k8s.Client
	ctx        context.Context
	nsx        *nsx.NSXClient
	kubeConfig []byte
	ca         certs.CertificateAuthority
	ingressCA  certs.CertificateAuthority
	// Terminating is true if the cluster is in a terminating state
	Terminating bool
}

func (platform *Platform) Init() {
	platform.Client.GetKubeConfigBytes = platform.GetKubeConfigBytes
	platform.Client.GetKustomizePatches = func() ([]string, error) {
		return platform.Patches, nil
	}
	platform.Client.ApplyDryRun = platform.DryRun
	platform.Client.Trace = platform.PlatformConfig.Trace
}

func (platform *Platform) GetKubeConfigBytes() ([]byte, error) {
	if platform.kubeConfig != nil {
		return platform.kubeConfig, nil
	}

	if platform.CA == nil || os.Getenv("KUBECONFIG") != "" {
		return []byte(files.SafeRead(os.Getenv("KUBECONFIG"))), nil
	}

	masters := platform.GetMasterIPs()
	if len(masters) == 0 {
		return nil, fmt.Errorf("could not find any master ips")
	}

	return k8s.CreateKubeConfig(platform.Name, platform.GetCA(), masters[0], "system:masters", "admin")
}

func (platform *Platform) GetCA() certs.CertificateAuthority {
	if platform.ca != nil {
		return platform.ca
	}
	ca, err := readCA(platform.CA)
	if err != nil {
		log.Fatalf("Unable to open %s: %v", platform.CA.PrivateKey, err)
	}
	platform.ca = ca
	return ca
}

func readCA(ca *types.CA) (*certs.Certificate, error) {
	cert := files.SafeRead(ca.Cert)
	privateKey := files.SafeRead(ca.PrivateKey)
	return certs.DecryptCertificate([]byte(cert), []byte(privateKey), []byte(ca.Password))
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
		log.Infof("Creating self-signed CA for ingress")
		ca := certs.NewCertificateBuilder("ingress-ca").CA().Certificate
		platform.ingressCA, _ = ca.SignCertificate(ca, 1)
		return platform.ingressCA
	}
	log.Debugf("[IngressCA] loading from disk: %s", platform.IngressCA.Cert)
	ca, err := readCA(platform.IngressCA)
	if err != nil {
		log.Fatalf("Unable to open Ingress CA: %v", err)
	}
	log.Debugf("[IngressCA] read CA %s", ca.X509.Subject)
	platform.ingressCA = ca
	return ca
}

// WaitFor at least 1 master IP to be reachable
func (platform *Platform) WaitFor() error {
	for {
		masters := platform.GetMasterIPs()
		if len(masters) > 0 && net.Ping(masters[0], 6443, 3) && platform.PingMaster() {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (platform *Platform) GetDNSClient() dns.Client {
	if platform.DNS == nil || platform.DNS.Disabled {
		return &dns.DummyDNSClient{Zone: platform.DNS.Zone}
	}

	if platform.DNS.Type == "route53" {
		dns := &dns.Route53Client{
			HostedZoneID: platform.DNS.Zone,
			AccessKey:    platform.DNS.AccessKey,
			SecretKey:    platform.DNS.SecretKey,
		}
		dns.Init()
		return dns
	}

	return &dns.DynamicDNSClient{
		Zone:       platform.DNS.Zone,
		KeyName:    platform.DNS.KeyName,
		Nameserver: platform.DNS.Nameserver,
		Key:        platform.DNS.Key,
		Algorithm:  platform.DNS.Algorithm,
	}
}

func (platform *Platform) GetNSXClient() (*nsx.NSXClient, error) {
	if platform.nsx != nil {
		return platform.nsx, nil
	}
	if platform.NSX == nil || platform.NSX.Disabled {
		return nil, fmt.Errorf("NSX not configured or disabled")
	}
	if platform.NSX.NsxV3 == nil || len(platform.NSX.NsxV3.NsxAPIManagers) == 0 {
		return nil, fmt.Errorf("nsx_v3.nsx_api_managers not configured")
	}

	client := &nsx.NSXClient{
		Host:     platform.NSX.NsxV3.NsxAPIManagers[0],
		Username: platform.NSX.NsxV3.NsxAPIUser,
		Password: platform.NSX.NsxV3.NsxAPIPass,
	}
	log.Debugf("Connecting to NSX-T %s@%s", client.Username, client.Host)

	if err := client.Init(); err != nil {
		return nil, fmt.Errorf("getNSXClient: failed to init client: %v", err)
	}
	platform.nsx = client
	version, err := platform.nsx.Ping()
	if err != nil {
		return nil, fmt.Errorf("getNSXClient: failed to ping: %v", err)
	}
	log.Infof("Logged into NSX-T %s@%s, version=%s", client.Username, client.Host, version)
	return platform.nsx, nil
}

func (platform *Platform) Clone(vm types.VM, config *konfigadm.Config) (types.Machine, error) {
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
		log.Warnf("Failed to set attributes for %s: %v", vm.Name, err)
	}
	log.Debugf("[%s] Waiting for IP", vm.Name)
	ip, err := VM.WaitForIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP for %s: %v", vm.Name, err)
	}
	vm.IP = ip
	log.Tracef("[%s] found ip %s", vm.Name, ip)
	if platform.NSX != nil && !platform.NSX.Disabled {
		ctx := context.TODO()
		nsxClient, err := platform.GetNSXClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get NSX client: %v", err)
		}

		ports, err := nsxClient.GetLogicalPorts(ctx, vm.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to find ports for %s: %v", vm.Name, err)
		}
		if len(ports) != 2 {
			return nil, fmt.Errorf("expected to find 2 ports, found %d", len(ports))
		}
		managementNic := make(map[string]string)
		transportNic := make(map[string]string)

		for k, v := range vm.Tags {
			managementNic[k] = v
		}

		transportNic["ncp/node_name"] = vm.Name
		transportNic["ncp/cluster"] = platform.Name

		if err := nsxClient.TagLogicalPort(ctx, ports[0].Id, managementNic); err != nil {
			return nil, fmt.Errorf("failed to tag management nic %s: %v", ports[0].Id, err)
		}
		if err := nsxClient.TagLogicalPort(ctx, ports[1].Id, transportNic); err != nil {
			return nil, fmt.Errorf("failed to tag transport nic %s: %v", ports[1].Id, err)
		}
	}
	return VM, nil
}

// GetMasterIPs returns a list of healthy master IP's
func (platform *Platform) GetMasterIPs() []string {
	if platform.Kubernetes.MasterIP != "" {
		return []string{platform.Kubernetes.MasterIP}
	}
	url := fmt.Sprintf("http://%s/v1/health/service/%s", platform.Consul, platform.Name)
	log.Tracef("Finding masters via consul: %s\n", url)
	response, _ := net.GET(url)
	var consul api.Consul
	if err := json.Unmarshal(response, &consul); err != nil {
		fmt.Println(err)
	}
	var addresses []string
node:
	for _, node := range consul {
		for _, check := range node.Checks {
			if check.Status != "passing" {
				log.Tracef("skipping unhealthy node %s -> %s", node.Node.Address, check.Status)
				continue node
			}
		}
		addresses = append(addresses, node.Node.Address)
	}
	return addresses
}

// GetKubeConfig gets the path to the admin kubeconfig, creating it if necessary
func (platform *Platform) GetKubeConfig() (string, error) {
	if os.Getenv("KUBECONFIG") != "" && os.Getenv("KUBECONFIG") != "false" {
		log.Tracef("Using KUBECONFIG from ENV")
		return os.Getenv("KUBECONFIG"), nil
	}
	cwd, _ := os.Getwd()
	name := cwd + "/" + platform.Name + "-admin.yml"
	if !is.File(name) {
		data, err := platform.GetKubeConfigBytes()
		if err != nil {
			return "", err
		}
		if err := ioutil.WriteFile(name, data, 0644); err != nil {
			return "", err
		}
	}
	return name, nil
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

	log.Tracef("Using KUBECONFIG=%s", kubeconfig)
	return deps.BinaryWithEnv("kubectl", platform.Kubernetes.Version, ".bin", map[string]string{
		"KUBECONFIG": kubeconfig,
		"PATH":       os.Getenv("PATH"),
	})
}

func (platform *Platform) CreateTLSSecret(namespace, subDomain, secretName string) error {
	if platform.HasSecret(namespace, secretName) {
		log.Debugf("secret/%s/%s' for %s alredy exists", namespace, secretName, subDomain)
		//TODO(moshloop) check certificate expiry and renew if necessary
		return nil
	}
	log.Infof("Creating new ingress cert %s.%s", subDomain, platform.Domain)
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
	log.Infof("Creating new ingress cert %s.%s", subDomain, platform.Domain)
	cert := certs.NewCertificateBuilder(subDomain + "." + platform.Domain).Server().Certificate
	return platform.GetIngressCA().SignCertificate(cert, 3)
}

func (platform *Platform) CreateInternalCertificate(service string, namespace string, clusterDomain string) (*certs.Certificate, error) {
	domain := fmt.Sprintf("%s.%s.svc.%s", service, namespace, clusterDomain)
	log.Infof("Creating new internal certificate %s", domain)
	cert := certs.NewCertificateBuilder(domain).Server().Certificate
	return platform.GetIngressCA().SignCertificate(cert, 5)
}

func (platform *Platform) GetResourceByName(file string, pkg string) (string, error) {
	var raw string
	var err error
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	if pkg == "manifests" {
		raw, err = manifests.FSString(false, file)
	} else {
		raw, err = templates.FSString(false, file)
	}
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
		log.Debugf("Error templating %s: %s", file, console.StripSecrets(string(data)))
		return "", err
	}
	return template, nil
}

func (platform *Platform) GetResourcesByDir(path string, pkg string) (map[string]http.File, error) {
	out := make(map[string]http.File)
	var fs http.FileSystem
	if pkg == "manifests" {
		fs = manifests.FS(false)
	} else {
		fs = templates.FS(false)
	}
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

func (platform *Platform) ExposeIngressTLS(namespace, service string, port int) error {
	return platform.Client.ExposeIngress(namespace, service, fmt.Sprintf("%s.%s", service, platform.Domain), port, map[string]string{
		"nginx.ingress.kubernetes.io/ssl-passthrough": "true",
	})
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

func (platform *Platform) ApplySpecs(namespace string, specs ...string) error {
	for _, spec := range specs {
		log.Debugf("Applying %s", spec)
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
			fmt.Printf("CMD: "+fmt.Sprintf("%s", env)+" .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	return deps.BinaryWithEnv(name, platform.Versions[name], ".bin", env)
}

func (platform *Platform) GetBinary(name string) deps.BinaryFunc {
	if platform.DryRun {
		return func(msg string, args ...interface{}) error {
			fmt.Printf("CMD: .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	return deps.Binary(name, platform.Versions[name], ".bin")
}

func (platform *Platform) GetOrCreateDB(name string, dbNames ...string) (*types.DB, error) {
	clusterName := "postgres-" + name
	databases := make(map[string]string)
	appUsername := "app"
	ns := "postgres-operator"
	secretName := fmt.Sprintf("%s.%s.credentials", appUsername, clusterName)

	db := &postgres.Postgresql{}
	if err := platform.Get(ns, clusterName, db); err != nil {
		log.Infof("Creating new cluster: %s", clusterName)
		for _, db := range dbNames {
			databases[db] = appUsername
		}
		db = postgres.NewPostgresql(clusterName)
		db.Spec.Databases = databases
		db.Spec.Users = map[string]postgres.UserFlags{
			appUsername: postgres.UserFlags{
				"createdb",
				"superuser",
			},
		}

		if err := platform.Apply(ns, db); err != nil {
			return nil, err
		}
	}

	doUntil(func() bool {
		if err := platform.Get(ns, clusterName, db); err != nil {
			return true
		}
		log.Infof("Waiting for %s to be running, is: %s", clusterName, db.Status.PostgresClusterStatus)
		return db.Status.PostgresClusterStatus == postgres.ClusterStatusRunning
	})
	if db.Status.PostgresClusterStatus != postgres.ClusterStatusRunning {
		return nil, fmt.Errorf("postgres cluster failed to start: %s", db.Status.PostgresClusterStatus)
	}
	secret := platform.GetSecret("postgres-operator", secretName)
	if secret == nil {
		return nil, fmt.Errorf("%s not found", secretName)
	}

	return &types.DB{
		Host:     fmt.Sprintf("%s.%s.svc.cluster.local", clusterName, ns),
		Username: string((*secret)["username"]),
		Port:     5432,
		Password: string((*secret)["password"]),
	}, nil
}

func doUntil(fn func() bool) bool {
	start := time.Now()

	for {
		if fn() {
			return true
		}
		if time.Now().After(start.Add(5 * time.Minute)) {
			return false
		}
		time.Sleep(5 * time.Second)
	}
}

func (platform *Platform) GetOrCreateBucket(name string) error {
	s3Client, err := platform.GetS3Client()
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %v", err)
	}

	exists, err := s3Client.BucketExists(name)
	if err != nil {
		return fmt.Errorf("failed to check S3 bucket: %v", err)
	}
	if !exists {
		log.Infof("Creating s3://%s", name)
		if err := s3Client.MakeBucket(name, platform.S3.Region); err != nil {
			return fmt.Errorf("failed to create S3 bucket: %v", err)
		}
	}
	return nil
}

func (platform *Platform) GetS3Client() (*minio.Client, error) {
	endpoint := platform.S3.GetExternalEndpoint()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	if endpoint == "" {
		endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", platform.S3.Region)
	}

	s3, err := minio.New(endpoint, platform.S3.AccessKey, platform.S3.SecretKey, false)
	if err != nil {
		return nil, err
	}
	s3.SetCustomTransport(client.Transport)
	return s3, nil
}
