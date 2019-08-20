package phases

import (
	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func NewClusterConfig(cfg *platform.Platform) api.ClusterConfiguration {
	cluster := api.ClusterConfiguration{
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

func NewInitConfig(cfg *platform.Platform) api.InitConfiguration {
	init := api.InitConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta1",
		Kind:       "InitConfiguration",
	}

	init.BootstrapTokens = []api.BootstrapToken{api.BootstrapToken{
		Groups: []string{"system:bootstrappers:kubeadm:default-node-token"},
		TTL:    "240h",
		Token:  cfg.BootstrapToken,
		Usages: []string{"signing", "authentication"},
	}}
	return init
}
