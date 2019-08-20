package api

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
