package types

import (
	"fmt"
	"github.com/moshloop/platform-cli/pkg/api/calico"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	ccm "k8s.io/kube-controller-manager/config/v1alpha1"
	proxy "k8s.io/kube-proxy/config/v1alpha1"
	scheduler "k8s.io/kube-scheduler/config/v1alpha1"
	kubelet "k8s.io/kubelet/config/v1beta1"
)

type PlatformConfig struct {
	Source               string            `yaml:"-"`
	ControlPlaneEndpoint string            `yaml:"-"`
	JoinEndpoint         string            `yaml:"-"`
	Certificates         *Certificates     `yaml:"-"`
	BuildOptions         BuildOptions      `yaml:"-"`
	Kubernetes           Kubernetes        `yaml:"kubernetes,omitempty"`
	BootstrapToken       string            `yaml:"token,omitempty"`
	Name                 string            `yaml:"name,omitempty"`
	Consul               string            `yaml:"consul,omitempty"`
	PodSubnet            string            `yaml:"podSubnet,omitempty"`
	ServiceSubnet        string            `yaml:"serviceSubnet,omitempty"`
	Calico               Calico            `yaml:"calico,omitempty"`
	DockerRegistry       string            `yaml:"dockerRegistry,omitempty"`
	Domain               string            `yaml:"domain,omitempty"`
	Ldap                 Ldap              `yaml:"ldap,omitempty"`
	SMTP                 Smtp              `yaml:"smtp,omitempty"`
	Specs                []string          `yaml:"specs,omitempty"`
	Policies             []string          `yaml:"policies,omitempty"`
	Monitoring           Monitoring        `yaml:"monitoring,omitempty"`
	ELK                  ELK               `yaml:"elk,omitempty"`
	Versions             map[string]string `yaml:"versions,omitempty"`
	Master               VM                `yaml:"master,omitempty"`
	Nodes                map[string]VM     `yaml:"workers,omitempty"`
	HostPrefix           string            `yaml:"hostPrefix,omitempty"`
	Harbor               Harbor            `yaml:"harbor,omitempty"`
	S3                   S3                `yaml:"s3,omitempty"`
}

type VM struct {
	Name         string `yaml:"name,omitempty"`
	Count        int    `yaml:"count,omitempty"`
	Template     string `yaml:"template,omitempty"`
	Cluster      string `yaml:"cluster,omitempty"`
	Folder       string `yaml:"folder,omitempty"`
	Datastore    string `yaml:"datastore,omitempty"`
	ResourcePool string `yaml:"resourcePool,omitempty"`
	CPUs         int32  `yaml:"cpu,omitempty"`
	MemoryGB     int64  `yaml:"memory,omitempty"`
	Network      string `yaml:"network,omitempty"`
	DiskGB       int    `yaml:"disk,omitempty"`
}

type Calico struct {
	IPIP      calico.IPIPMode         `yaml:"ipip"`
	VxLAN     calico.VXLANMode        `yaml:"vxlan"`
	Version   string                  `yaml:"version,omitempty"`
	Log       string                  `yaml:"log,omitempty"`
	BGPPeers  []calico.BGPPeer        `yaml:"bgpPeers,omitempty"`
	BGPConfig calico.BGPConfiguration `yaml:"bgpConfig,omitempty"`
	IPPools   []calico.IPPool         `yaml:"ipPools,omitempty"`
}

type Harbor struct {
	Version string `yaml:"version,omitempty"`
}

type Certificates struct {
	OpenID     Certificate `yaml:"open_id,omitempty"`
	Etcd       Certificate `yaml:"etcd,omitempty"`
	FrontProxy Certificate `yaml:"front_proxy,omitempty"`
	SA         Certificate `yaml:"sa,omitempty"`
	CA         Certificate `yaml:"ca,omitempty"`
}
type Certificate struct {
	Key  string `yaml:"key,omitempty"`
	X509 string `yaml:"x509,omitempty"`
}

func (c Certificate) ToCert() *utils.Certificate {
	cert, _ := utils.DecodeCertificate([]byte(c.X509), []byte(c.Key))
	return cert
}

type Smtp struct {
	Server   string `yaml:"server,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	From     string `yaml:"from,omitempty"`
}

type S3 struct {
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Bucket    string `yaml:"bucket,omitempty"`
	Region    string `yaml:"region,omitempty"`
	Endpoint  string `yaml:"endpoint,omitempty"`
}

type Ldap struct {
	Host       string `yaml:"host,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	Domain     string `yaml:"domain,omitempty"`
	AdminGroup string `yaml:"adminGroup,omitempty"`
	BindDN     string `yaml:"dn,omitempty"`
}
type BuildOptions struct {
	Monitoring bool
}

type HelmChart struct {
	Repo       string            `yaml:"repo,omitempty"`
	Chart      string            `yaml:"chart,omitempty"`
	Version    string            `yaml:"version,omitempty"`
	Values     map[string]string `yaml:"values,omitempty"`
	ValuesFile string            `yaml:"valuesFile,omitempty"`
}

type Kubernetes struct {
	Version           string                       `yaml:"version,omitempty"`
	Kubelet           kubelet.KubeletConfiguration `yaml:"kubelet,omitempty"`
	KubeProxy         proxy.KubeProxyConfiguration
	KubeScheduler     scheduler.KubeSchedulerConfiguration
	ControllerManager ccm.KubeControllerManagerConfiguration
}

type ObjectStorage struct {
	AccessKey    string `yaml:"accessKey,omitempty"`
	SecretKey    string `yaml:"secretKey,omitempty"`
	Endpoint     string `yaml:"endpoint,omitempty"`
	Bucket       string `yaml:"bucket,omitempty"`
	RegistryPath string `yaml:"registry_path,omitempty"`
}
type Monitoring struct {
	Prometheus Prometheus `yaml:"prometheus,omitempty"`
	Grafana    Grafana    `yaml:"grafana,omitempty"`
}

type Prometheus struct {
	Version string `yaml:"version,omitempty"`
}

type Grafana struct {
	Version string `yaml:"version,omitempty"`
}

type ELK struct {
	Version      string `yaml:"version,omitempty"`
	Replicas     int    `yaml:"replicas,omitempty"`
	LogRetention string `yaml:"logRetention,omitempty"`
}

type Versions struct {
	Kubernetes       string            `yaml:"kubernetes,omitempty"`
	ContainerRuntime string            `yaml:"containerRuntime,omitempty"`
	Dependencies     map[string]string `yaml:"dependencies,omitempty"`
}

func (p PlatformConfig) GetImagePath(image string) string {
	if p.DockerRegistry == "" {
		return image
	}
	return p.DockerRegistry + "/" + image
}

func (p PlatformConfig) GetVMCount() int {
	count := p.Master.Count
	for _, node := range p.Nodes {
		count += node.Count
	}
	return count
}

func (platform *PlatformConfig) Init() {
	if platform.BootstrapToken == "" {
		platform.BootstrapToken = GenerateBootstrapToken()
		log.Infof("Created new bootstrap token %s\n", platform.BootstrapToken)
	}
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}
	if platform.Certificates == nil {
		platform.Certificates = GetCertificates(*platform)
	}
}

// GenerateBootstrapToken generates a new kubeadm bootstrap token
func GenerateBootstrapToken() string {
	return fmt.Sprintf("%s.%s", utils.RandomString(6), utils.RandomString(16))
}

func GenerateCA(name string) Certificate {
	cert, _ := utils.NewCertificateAuthority(name)
	return Certificate{
		Key:  string(cert.EncodedPrivateKey()),
		X509: string(cert.EncodedCertificate()),
	}
}

func GetCertificates(platform PlatformConfig) *Certificates {
	file := platform.Name + "_cert.yaml"
	if utils.FileExists(file) {
		var certs Certificates
		data, _ := ioutil.ReadFile(file)
		yaml.Unmarshal(data, &certs)
		log.Infof("Loaded certificates from %s\n", file)
		return &certs
	}

	log.Infoln("Generating certificates")

	certs := Certificates{
		Etcd:       GenerateCA("etcd-ca"),
		FrontProxy: GenerateCA("front-proxy-ca"),
		CA:         GenerateCA("kubernetes"),
		SA:         GenerateCA("sa-ca"),
		OpenID:     GenerateCA("dex." + platform.Domain),
	}

	data, _ := yaml.Marshal(certs)
	ioutil.WriteFile(file, data, 0644)
	log.Infof("Saved certificates to %s\n", file)
	return &certs
}
