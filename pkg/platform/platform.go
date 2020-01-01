package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/moshloop/commons/console"
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/is"
	"github.com/moshloop/commons/net"
	"github.com/moshloop/commons/text"
	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/manifests"
	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/client/dns"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/nsx"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
)

type Platform struct {
	types.PlatformConfig
	k8s.Client
	ctx     context.Context
	session *vmware.Session
	nsx     *nsx.NSXClient
}

func (platform *Platform) Init() {
	platform.Client.GetKubeConfig = platform.GetKubeConfig
	platform.Client.ApplyDryRun = platform.DryRun
}

// GetVMs returns a list of all VM's associated with the cluster
func (platform *Platform) GetVMsByPrefix(prefix string) (map[string]*VM, error) {
	var vms = make(map[string]*VM)
	list, err := platform.session.Finder.VirtualMachineList(
		platform.ctx, fmt.Sprintf("%s-%s-%s*", platform.HostPrefix, platform.Name, prefix))
	if err != nil {
		return nil, nil
	}
	for _, vm := range list {
		item := &VM{
			Platform: platform,
			ctx:      platform.ctx,
			vm:       vm,
		}
		item.Name = vm.Name()
		vms[vm.Name()] = item
	}
	return vms, nil
}

// GetVMs returns a list of all VM's associated with the cluster
func (platform *Platform) GetVMs() (map[string]*VM, error) {
	return platform.GetVMsByPrefix("")
}

// WaitFor at least 1 master IP to be reachable
func (platform *Platform) WaitFor() error {
	for {
		masters := platform.GetMasterIPs()
		if len(masters) > 0 && net.Ping(masters[0], 6443, 3) {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (platform *Platform) GetDNSClient() dns.DNSClient {
	if platform.DNS == nil || platform.DNS.Disabled {
		return dns.DummyDNSClient{Zone: platform.DNS.Zone}
	}
	return dns.DynamicDNSClient{
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
	if platform.NSX.NsxV3 == nil || len(platform.NSX.NsxV3.NsxApiManagers) == 0 {
		return nil, fmt.Errorf("nsx_v3.nsx_api_managers not configured")
	}

	client := &nsx.NSXClient{
		Host:     platform.NSX.NsxV3.NsxApiManagers[0],
		Username: platform.NSX.NsxV3.NsxApiUser,
		Password: platform.NSX.NsxV3.NsxApiPass,
	}
	log.Debugf("Connecting to NSX-T %s@%s", client.Username, client.Host)

	if err := client.Init(); err != nil {
		return nil, err
	}
	platform.nsx = client
	version, err := platform.nsx.Ping()
	if err != nil {
		return nil, err
	}
	log.Infof("Logged into NSX-T %s@%s, version=%s", client.Username, client.Host, version)
	return platform.nsx, nil
}

func (platform *Platform) Clone(vm types.VM, config *konfigadm.Config) (*VM, error) {
	for _, cmd := range vm.Commands {
		config.AddCommand(cmd)
	}
	ctx := context.TODO()
	obj, err := platform.session.Clone(vm, config)
	if err != nil {
		return nil, err
	}

	VM := &VM{
		VM:       vm,
		Platform: platform,
		ctx:      platform.ctx,
		vm:       obj,
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
		return nil, fmt.Errorf("Failed to get IP for %s: %v", vm.Name, err)
	}
	VM.IP = ip
	if platform.NSX != nil && !platform.NSX.Disabled {
		nsxClient, err := platform.GetNSXClient()
		if err != nil {
			return nil, fmt.Errorf("Failed to get NSX client: %v", err)
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

func (platform *Platform) GetSession() *vmware.Session {
	return platform.session
}

// OpenViaEnv opens a new vmware session using environment variables
func (platform *Platform) OpenViaEnv() error {
	if platform.session != nil {
		return nil
	}
	platform.ctx = context.TODO()
	session, err := vmware.GetSessionFromEnv()
	if err != nil {
		return err
	}
	platform.session = session
	return nil
}

// GetMasterIPs returns a list of healthy master IP's
func (platform *Platform) GetMasterIPs() []string {
	url := fmt.Sprintf("http://%s/v1/health/service/%s", platform.Consul, platform.Name)
	log.Infof("Finding masters via consul: %s\n", url)
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
		log.Tracef("Using KUBECONFIG from ENV\n")
		return os.Getenv("KUBECONFIG"), nil
	}
	cwd, _ := os.Getwd()
	name := cwd + "/" + platform.Name + "-admin.yml"
	if !is.File(name) {
		masters := platform.GetMasterIPs()
		if len(masters) == 0 {
			return "", errors.New("No masters found")
		}
		data, err := CreateKubeConfig(platform, masters[0])
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
			return fmt.Errorf("cannot create kubeconfig %v\n", err)
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
			return fmt.Errorf("cannot create kubeconfig %v\n", err)
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

// GetSecret returns the data of a secret or nil for any error
func (platform *Platform) GetSecret(namespace, name string) *map[string][]byte {
	k8s, err := platform.GetClientset()
	if err != nil {
		log.Tracef("Failed to get client %v", err)
		return nil
	}
	secret, err := k8s.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Tracef("Failed to get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &secret.Data
}

// GetSecret returns the data of a secret or nil for any error
func (platform *Platform) GetConfigMap(namespace, name string) *map[string]string {
	k8s, err := platform.GetClientset()
	if err != nil {
		log.Tracef("Failed to get client %v", err)
		return nil
	}
	cm, err := k8s.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Tracef("Failed tp get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &cm.Data
}

// CreateKubeConfig creates a new kubeconfig for the cluster
func CreateKubeConfig(platform *Platform, endpoint string) ([]byte, error) {
	userName := "kubernetes-admin"
	contextName := fmt.Sprintf("%s@%s", userName, platform.Name)
	cert, err := platform.Certificates.CA.ToCert().CreateCertificate("system:masters", "system:masters")
	if err != nil {
		return nil, err
	}
	cfg := kapi.Config{
		Clusters: map[string]*kapi.Cluster{
			platform.Name: {
				Server:                   "https://" + endpoint + ":6443",
				CertificateAuthorityData: []byte(platform.Certificates.CA.X509),
			},
		},
		Contexts: map[string]*kapi.Context{
			contextName: {
				Cluster:  platform.Name,
				AuthInfo: userName,
			},
		},
		AuthInfos: map[string]*kapi.AuthInfo{
			userName: {
				ClientKeyData:         cert.EncodedPrivateKey(),
				ClientCertificateData: cert.EncodedCertificate(),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.Write(cfg)
}

// GetDynamicClient creates a new k8s client
func (platform *Platform) GetDynamicClient() (dynamic.Interface, error) {
	kubeconfig, err := platform.GetKubeConfig()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

// GetClientset creates a new k8s client
func (platform *Platform) GetClientset() (*kubernetes.Clientset, error) {
	kubeconfig, err := platform.GetKubeConfig()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfigOrDie(cfg), nil
}

func (platform *Platform) GetResourceByName(file string) (string, error) {
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	raw, err := manifests.FSString(false, file)
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (platform *Platform) Template(file string) (string, error) {
	raw, err := platform.GetResourceByName(file)
	if err != nil {
		return "", fmt.Errorf("Could not find %s: %v", file, err)
	}
	if strings.HasSuffix(file, ".raw") {
		return raw, nil
	}
	template, err := text.Template(raw, platform.PlatformConfig)
	if err != nil {
		data, _ := yaml.Marshal(platform.PlatformConfig)
		log.Debugln("Error templating %s: %s", file, console.StripSecrets(string(data)))
		return "", err
	}
	return template, nil
}

func (platform *Platform) GetResourcesByDir(path string) (map[string]http.File, error) {
	out := make(map[string]http.File)
	fs := manifests.FS(false)

	dir, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		file, err := fs.Open(path + "/" + info.Name())
		if err != nil {
			return nil, err
		}
		out[info.Name()] = file
	}
	return out, nil
}

func (platform *Platform) TemplateDir(path string) (string, error) {
	fs := manifests.FS(false)

	tmp, _ := ioutil.TempDir("", "template")

	dir, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return "", err
	}
	dst := ".manifests/" + path
	for _, info := range files {
		to := tmp + "/" + info.Name()
		if strings.HasSuffix(info.Name(), ".raw") {
			to = dst + "/" + info.Name()
		}
		log.Debugf("Extracting %s\n", to)
		destination, err := os.Create(to)
		if err != nil {
			return "", err
		}
		defer destination.Close()
		file, err := fs.Open(path + "/" + info.Name())
		if err != nil {
			return "", err
		}
		_, err = io.Copy(destination, file)
		if err != nil {
			return "", err
		}
	}
	os.RemoveAll(dst)
	os.MkdirAll(dst, 0775)
	return dst, text.TemplateDir(tmp, dst, platform.PlatformConfig)
}

func (platform *Platform) Annotate(objectType, name, namespace string, annotations map[string]string) error {
	if len(annotations) == 0 {
		return nil
	}
	kubectl := platform.GetKubectl()
	if namespace != "" {
		namespace = "-n " + namespace
	}

	var (
		line  string
		lines []string
	)

	for k, v := range annotations {
		line = fmt.Sprintf("%s=\"%s\"", k, v)
		lines = append(lines, line)
	}

	return kubectl("annotate %s %s %s %s", objectType, name, strings.Join(lines, " "), namespace)
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
	kubectl := platform.GetKubectl()
	if namespace != "" {
		namespace = "-n " + namespace
	}

	for _, spec := range specs {
		data, err := yaml.Marshal(spec)
		if err != nil {
			return err
		}

		if log.IsLevelEnabled(log.TraceLevel) {
			log.Tracef("Applying %s/%s/%s:\n%s", namespace, spec.Kind, spec.Metadata.Name, string(data))
		} else {
			log.Debugf("Applying  %s/%s/%s", namespace, spec.Kind, spec.Metadata.Name)
		}

		file := text.ToFile(string(data), ".yml")
		if err := kubectl("apply %s -f %s", namespace, file); err != nil {
			return err
		}
	}
	return nil
}

func (platform *Platform) ApplyText(namespace string, specs ...string) error {
	kubectl := platform.GetKubectl()
	if namespace != "" {
		namespace = "-n " + namespace
	}

	for _, spec := range specs {
		file := text.ToFile(spec, ".yml")
		if err := kubectl("apply %s -f %s", namespace, file); err != nil {
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
	kubectl := platform.GetKubectl()
	if namespace != "" {
		namespace = "-n " + namespace
	}
	for _, spec := range specs {
		if strings.HasSuffix(spec, "/") {
			dir, err := platform.TemplateDir(spec)
			if err != nil {
				return err
			}
			if err := kubectl("apply %s -f %s", namespace, dir); err != nil {
				return err
			}
		} else {
			template, err := platform.Template(spec)
			if err != nil {
				return err
			}
			if err := kubectl("apply %s -f %s", namespace, text.ToFile(template, ".yaml")); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Platform) GetBinaryWithEnv(name string, env map[string]string) deps.BinaryFunc {
	if p.DryRun {
		return func(msg string, args ...interface{}) error {
			fmt.Printf("CMD: "+fmt.Sprintf("%s", env)+" .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	return deps.BinaryWithEnv(name, p.Versions[name], ".bin", env)
}

func (p *Platform) GetBinary(name string) deps.BinaryFunc {
	if p.DryRun {
		return func(msg string, args ...interface{}) error {
			fmt.Printf("CMD: .bin/"+name+" "+msg+"\n", args...)
			return nil
		}
	}
	return deps.Binary(name, p.Versions[name], ".bin")
}
