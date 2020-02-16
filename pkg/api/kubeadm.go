package api

type ClusterConfiguration struct {
	APIVersion           string `yaml:"apiVersion,omitempty"`
	Kind                 string `yaml:"kind"`
	KubernetesVersion    string `yaml:"kubernetesVersion,omitempty"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint,omitempty"`
	APIServer            struct {
		CertSANs               []string          `yaml:"certSANs,omitempty"`
		TimeoutForControlPlane string            `yaml:"timeoutForControlPlane,omitempty"`
		ExtraArgs              map[string]string `yaml:"extraArgs,omitempty"`
	} `yaml:"apiServer,omitempty"`
	CertificatesDir   string `yaml:"certificatesDir,omitempty"`
	ClusterName       string `yaml:"clusterName,omitempty"`
	ControllerManager struct {
		ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
	} `yaml:"controllerManager,omitempty"`
	DNS struct {
		Type string `yaml:"type,omitempty"`
	} `yaml:"dns,omitempty"`
	Etcd struct {
		Local struct {
			DataDir   string            `yaml:"dataDir,omitempty"`
			ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
		} `yaml:"local,omitempty"`
	} `yaml:"etcd,omitempty"`
	ImageRepository string `yaml:"imageRepository,omitempty"`
	Networking      struct {
		DNSDomain     string `yaml:"dnsDomain,omitempty"`
		ServiceSubnet string `yaml:"serviceSubnet,omitempty"`
		PodSubnet     string `yaml:"podSubnet,omitempty"`
	} `yaml:"networking,omitempty"`
	Scheduler struct {
		ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
	} `yaml:"scheduler,omitempty"`
}

type InitConfiguration struct {
	APIVersion       string           `yaml:"apiVersion,omitempty"`
	Kind             string           `yaml:"kind"`
	BootstrapTokens  []BootstrapToken `yaml:"bootstrapTokens,omitempty"`
	NodeRegistration NodeRegistration `yaml:"nodeRegistration,omitempty"`
}

type BootstrapToken struct {
	Groups []string `yaml:"groups"`
	Token  string   `yaml:"token"`
	TTL    string   `yaml:"ttl"`
	Usages []string `yaml:"usages"`
}

type NodeRegistration struct {
	KubeletExtraArgs map[string]string `yaml:"kubeletExtraArgs,omitempty"`
}
