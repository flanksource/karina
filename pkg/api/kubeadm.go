package api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type HostPathType string

const (
	// For backwards compatible, leave it empty if unset
	HostPathUnset HostPathType = ""
	// If nothing exists at the given path, an empty directory will be created there
	// as needed with file mode 0755, having the same group and ownership with Kubelet.
	HostPathDirectoryOrCreate HostPathType = "DirectoryOrCreate"
	// A directory must exist at the given path
	HostPathDirectory HostPathType = "Directory"
	// If nothing exists at the given path, an empty file will be created there
	// as needed with file mode 0644, having the same group and ownership with Kubelet.
	HostPathFileOrCreate HostPathType = "FileOrCreate"
	// A file must exist at the given path
	HostPathFile HostPathType = "File"
	// A UNIX socket must exist at the given path
	HostPathSocket HostPathType = "Socket"
	// A character device must exist at the given path
	HostPathCharDev HostPathType = "CharDevice"
	// A block device must exist at the given path
	HostPathBlockDev HostPathType = "BlockDevice"
)

type ClusterConfiguration struct {
	APIVersion           string `yaml:"apiVersion,omitempty"`
	Kind                 string `yaml:"kind"`
	KubernetesVersion    string `yaml:"kubernetesVersion,omitempty"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint,omitempty"`
	APIServer            struct {
		CertSANs               []string          `yaml:"certSANs,omitempty"`
		TimeoutForControlPlane string            `yaml:"timeoutForControlPlane,omitempty"`
		ExtraArgs              map[string]string `yaml:"extraArgs,omitempty"`
		ExtraVolumes           []HostPathMount   `yaml:"extraVolumes,omitempty"`
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

type JoinConfiguration struct {
	APIVersion       string           `yaml:"apiVersion,omitempty"`
	Kind             string           `yaml:"kind"`
	NodeRegistration NodeRegistration `yaml:"nodeRegistration,omitempty"`

	// Discovery specifies the options for the kubelet to use during the TLS Bootstrap process
	Discovery Discovery `yaml:"discovery"`

	// ControlPlane defines the additional control plane instance to be deployed on the joining node.
	// If nil, no additional control plane instance will be deployed.
	ControlPlane *JoinControlPlane `yaml:"controlPlane,omitempty"`
}

type JoinControlPlane struct {
	// LocalAPIEndpoint represents the endpoint of the API server instance to be deployed on this node.
	LocalAPIEndpoint APIEndpoint `yaml:"localAPIEndpoint,omitempty"`

	// CertificateKey is the key that is used for decryption of certificates after they are downloaded from the secret
	// upon joining a new control plane node. The corresponding encryption key is in the InitConfiguration.
	CertificateKey string `yaml:"certificateKey,omitempty"`
}

//nolint: revive
type APIEndpoint struct {
	// AdvertiseAddress sets the IP address for the API server to advertise.
	AdvertiseAddress string `yaml:"advertiseAddress,omitempty"`

	// BindPort sets the secure port for the API Server to bind to.
	// Defaults to 6443.
	BindPort int32 `yaml:"bindPort,omitempty"`
}

type Discovery struct {
	// BootstrapToken is used to set the options for bootstrap token based discovery
	// BootstrapToken and File are mutually exclusive
	BootstrapToken *BootstrapTokenDiscovery `yaml:"bootstrapToken,omitempty"`

	// Timeout modifies the discovery timeout
	Timeout *metav1.Duration `yaml:"timeout,omitempty"`
}

type BootstrapTokenDiscovery struct {
	// Token is a token used to validate cluster information
	// fetched from the control-plane.
	Token string `json:"token"`

	// APIServerEndpoint is an IP or domain name to the API server from which info will be fetched.
	APIServerEndpoint string `yaml:"apiServerEndpoint,omitempty"`

	// CACertHashes specifies a set of public key pins to verify
	// when token-based discovery is used. The root CA found during discovery
	// must match one of these values. Specifying an empty set disables root CA
	// pinning, which can be unsafe. Each hash is specified as "<type>:<value>",
	// where the only currently supported type is "sha256". This is a hex-encoded
	// SHA-256 hash of the Subject Public Key Info (SPKI) object in DER-encoded
	// ASN.1. These hashes can be calculated using, for example, OpenSSL.
	CACertHashes []string `yaml:"caCertHashes,omitempty"`

	// UnsafeSkipCAVerification allows token-based discovery
	// without CA verification via CACertHashes. This can weaken
	// the security of kubeadm since other nodes can impersonate the control-plane.
	UnsafeSkipCAVerification bool `yaml:"unsafeSkipCAVerification,omitempty"`
}

type BootstrapToken struct {
	Groups []string `yaml:"groups"`
	Token  string   `yaml:"token"`
	TTL    string   `yaml:"ttl"`
	Usages []string `yaml:"usages"`
}

type NodeRegistration struct {
	CRISocket        string            `yaml:"criSocket,omitempty"`
	KubeletExtraArgs map[string]string `yaml:"kubeletExtraArgs,omitempty"`
}

type HostPathMount struct {
	// Name of the volume inside the pod template.
	Name string `yaml:"name"`
	// HostPath is the path in the host that will be mounted inside
	// the pod.
	HostPath string `yaml:"hostPath"`
	// MountPath is the path inside the pod where hostPath will be mounted.
	MountPath string `yaml:"mountPath"`
	// ReadOnly controls write access to the volume
	ReadOnly bool `yaml:"readOnly,omitempty"`
	// PathType is the type of the HostPath.
	PathType HostPathType `yaml:"pathType,omitempty"`
}
