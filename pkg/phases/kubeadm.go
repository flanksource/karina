package phases

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func NewClusterConfig(cfg *platform.Platform) ClusterConfiguration {
	cluster := ClusterConfiguration{
		APIVersion:        "kubeadm.k8s.io/v1beta2",
		Kind:              "ClusterConfiguration",
		KubernetesVersion: cfg.Versions.Kubernetes,
		CertificatesDir:   "/etc/kubernetes/pki",
		ClusterName:       cfg.Name,
		ImageRepository:   "k8s.gcr.io",
		// Control plane endpoint is load balanced client side using haproxy + consul service discovery
		ControlPlaneEndpoint: "localhost:8443",
	}
	cluster.Networking.DNSDomain = "cluster.local"
	cluster.Networking.ServiceSubnet = cfg.ServiceSubnet
	cluster.Networking.PodSubnet = cfg.PodSubnet
	cluster.DNS.Type = "CoreDNS"
	cluster.Etcd.Local.DataDir = "/var/lib/etcd"
	cluster.APIServer.CertSANs = []string{"localhost", "127.0.0.1"}
	cluster.APIServer.TimeoutForControlPlane = "4m0s"
	cluster.APIServer.ExtraArgs = map[string]string{
		"oidc-issuer-url":     "https://dex." + cfg.Domain,
		"oidc-client-id":      "kubernetes",
		"oidc-ca-file":        "/etc/ssl/certs/openid-ca.pem",
		"oidc-username-claim": "email",
		"oidc-groups-claim":   "groups",
	}
	return cluster
}

func NewInitConfig(cfg *platform.Platform) InitConfiguration {
	init := InitConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta1",
		Kind:       "InitConfiguration",
	}

	init.BootstrapTokens = []BootstrapToken{BootstrapToken{
		Groups: []string{"system:bootstrappers:kubeadm:default-node-token"},
		TTL:    "240h",
		Token:  cfg.BootstrapToken,
		Usages: []string{"signing", "authentication"},
	}}
	return init
}

type ClusterConfiguration struct {
	APIVersion           string `yaml:"apiVersion"`
	Kind                 string `yaml:"kind"`
	KubernetesVersion    string `yaml:"kubernetesVersion"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint,omitempty"`
	APIServer            struct {
		CertSANs               []string          `yaml:"certSANs,omitempty"`
		TimeoutForControlPlane string            `yaml:"timeoutForControlPlane"`
		ExtraArgs              map[string]string `yaml:"extraArgs"`
	} `yaml:"apiServer"`
	CertificatesDir   string `yaml:"certificatesDir"`
	ClusterName       string `yaml:"clusterName"`
	ControllerManager struct {
		ExtraArgs map[string]string `yaml:"extraArgs"`
	} `yaml:"controllerManager"`
	DNS struct {
		Type string `yaml:"type"`
	} `yaml:"dns"`
	Etcd struct {
		Local struct {
			DataDir string `yaml:"dataDir"`
		} `yaml:"local"`
	} `yaml:"etcd"`
	ImageRepository string `yaml:"imageRepository"`
	Networking      struct {
		DNSDomain     string `yaml:"dnsDomain"`
		ServiceSubnet string `yaml:"serviceSubnet"`
		PodSubnet     string `yaml:"podSubnet"`
	} `yaml:"networking"`
	Scheduler struct {
		ExtraArgs map[string]string `yaml:"extraArgs"`
	} `yaml:"scheduler"`
}

type InitConfiguration struct {
	APIVersion      string           `yaml:"apiVersion"`
	Kind            string           `yaml:"kind"`
	BootstrapTokens []BootstrapToken `yaml:"bootstrapTokens"`
}

type BootstrapToken struct {
	Groups []string `yaml:"groups"`
	Token  string   `yaml:"token"`
	TTL    string   `yaml:"ttl"`
	Usages []string `yaml:"usages"`
}
