package types

import (
	"github.com/moshloop/platform-cli/pkg/utils"
)

type PlatformConfig struct {
	Source               string        `yaml:"-"`
	ControlPlaneEndpoint string        `yaml:"-"`
	JoinEndpoint         string        `yaml:"-"`
	Certificates         *Certificates `yaml:"-"`
	BuildOptions         BuildOptions  `yaml:"-"`
	BootstrapToken       string        `yaml:"token,omitempty"`
	Name                 string        `yaml:"name,omitempty"`
	Consul               string        `yaml:"consul,omitempty"`
	PodSubnet            string        `yaml:"podSubnet,omitempty"`
	ServiceSubnet        string        `yaml:"serviceSubnet,omitempty"`
	DockerRegistry       string        `yaml:"dockerRegistry,omitempty"`
	Domain               string        `yaml:"domain,omitempty"`
	Ldap                 Ldap          `yaml:"ldap,omitempty"`
	SMTP                 Smtp          `yaml:"smtp,omitempty"`
	Specs                []string      `yaml:"specs,omitempty"`
	Policies             []string      `yaml:"policies,omitempty"`
	Monitoring           Monitoring    `yaml:"monitoring,omitempty"`
	ELK                  ELK           `yaml:"elk,omitempty"`
	Versions             Versions      `yaml:"versions,omitempty"`
	Master               VM            `yaml:"master,omitempty"`
	Nodes                map[string]VM `yaml:"workers,omitempty"`
	HostPrefix           string        `yaml:"hostPrefix,omitempty"`
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

type ObjectStorage struct {
	AccessKey string `yaml:"accessKey,omitempty"`
	SecretKey string `yaml:"secretKey,omitempty"`
	Endpoint  string `yaml:"endpoint,omitempty"`
	Bucket    string `yaml:"bucket,omitempty"`
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
