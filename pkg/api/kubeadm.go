package api

import (
	"github.com/moshloop/platform-cli/pkg/types"
	"strconv"
)

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

type BootstrapToken struct {
	Groups []string `yaml:"groups"`
	Token  string   `yaml:"token"`
	TTL    string   `yaml:"ttl"`
	Usages []string `yaml:"usages"`
}

type NodeRegistration struct {
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

// Set the APIServer ExtraArguments related to Audit logging from a supplied
// github.com/moshloop/platform-cli/pkg/types/ApiServerOptions
func (c *ClusterConfiguration) SetAPIServerExtraAuditArgs(logOptions types.ApiServerOptions) {
	c.setAPIServerExtraAuditLogArgs(logOptions.LogOptions)
	c.setAPIServerExtraAuditWebhookArgs(logOptions.WebhookOptions)
}

// Internal helper function to set the APIServer ExtraArguments related to Audit logging from a supplied
// github.com/moshloop/platform-cli/pkg/types/AuditLogOptions
// These options are for Audit logfiles only
func (c *ClusterConfiguration) setAPIServerExtraAuditLogArgs(logOptions types.AuditLogOptions)  {
	// Options for logfiles
	if logOptions.Path != "" {
		c.APIServer.ExtraArgs["audit-log-path"] = logOptions.Path
	}

	if logOptions.Format != "" {
		c.APIServer.ExtraArgs["audit-log-format"] = logOptions.Format
	}
	if logOptions.MaxAge != 0 {
		c.APIServer.ExtraArgs["audit-log-maxage"] = strconv.Itoa(logOptions.MaxAge)
	}
	if logOptions.MaxBackups != 0 {
		c.APIServer.ExtraArgs["audit-log-maxbackup"] = strconv.Itoa(logOptions.MaxBackups)
	}
	if logOptions.MaxSize != 0 {
		c.APIServer.ExtraArgs["audit-log-maxsize"] = strconv.Itoa(logOptions.MaxSize)
	}
}

// Internal helper function to set the APIServer ExtraArguments related to Audit logging from a supplied
// github.com/moshloop/platform-cli/pkg/types/AuditWebhookOptions
// These options are for Audit Webhooks only
func (c *ClusterConfiguration) setAPIServerExtraAuditWebhookArgs(logOptions types.AuditWebhookOptions)  {
	// Options for webhooks
	if logOptions.ConfigFile != "" {
		c.APIServer.ExtraArgs["audit-webhook-config-file"] = logOptions.ConfigFile
	}
	if logOptions.InitialBackoff != 0 {
		c.APIServer.ExtraArgs["audit-webhook-initial-backoff"] = logOptions.InitialBackoff.String()
	}

}

