package phases

import (
	"fmt"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/logger"
	"gopkg.in/flanksource/yaml.v3"

	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	_ "github.com/flanksource/konfigadm/pkg" // initialize konfigadm
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/pkg/errors"
)

var envVars = map[string]string{
	"ETCDCTL_ENDPOINTS": "https://127.0.0.1:2379",
	"ETCDCTL_CACERT":    "/etc/kubernetes/pki/etcd/ca.crt",
	"ETCDCTL_CERT":      "/etc/kubernetes/pki/etcd/healthcheck-client.crt",
	"ETCDCTL_KEY":       "/etc/kubernetes/pki/etcd/healthcheck-client.key",
	"KUBECONFIG":        "/etc/kubernetes/admin.conf",
}

const (
	updateHostsFileCmd = "echo $(ifconfig ens160 | grep inet | awk '{print $2}' | head -n1 ) $(hostname) >> /etc/hosts"
	kubeadmInitCmd     = "kubeadm init --config /etc/kubernetes/kubeadm.conf -v 5 2>&1 | tee -a /var/log/kubeadm.log"
	kubeadmNodeJoinCmd = "kubeadm join --config /etc/kubernetes/kubeadm.conf -v 5 2>&1 | tee -a /var/log/kubeadm.log"
)

var downloadCustomClusterSigningFiles = []string{
	// Because the certificate-signing-cert can only contain a single certificate and the ca.crt contains 2 (the root ca, and cluster ca)
	// ca.{key,crt} are copied removing the 2nd cert and specified as extra arguments to the api server, because of this kubeadm upload/download certs
	// is not aware of these and they get recreated for each new master, causing node join issues down the line, the fix below is to recreate these certs
	// from the ca cert|key downloaded by kubeadm
	"kubeadm join phase control-plane-prepare download-certs --config /etc/kubernetes/kubeadm.conf -v 5 2>&1 | tee -a /var/log/kubeadm.log",
	"cp /etc/kubernetes/pki/ca.key /etc/kubernetes/pki/csr-ca.key",
	"cat /etc/kubernetes/pki/ca.crt | awk '1;/-----END CERTIFICATE-----/{exit}' > /etc/kubernetes/pki/csr-ca.crt",
}

// CreatePrimaryMaster creates a konfigadm config for the primary master.
func CreatePrimaryMaster(platform *platform.Platform) (*konfigadm.Config, error) {
	if platform.Name == "" {
		return nil, errors.New("Must specify a platform name")
	}
	cfg, err := baseKonfig(platform.Master.KonfigadmFile, platform)
	if err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to get baseKonfig: %v", err)
	}
	if err := addInitKubeadmConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to add kubeadm config: %v", err)
	}
	files, err := kubeadm.GetFilesToMountForPrimary(platform)
	if err != nil {
		return nil, err
	}
	for file, content := range files {
		cfg.Files[file] = content
	}
	cfg.AddCommand(kubeadmInitCmd)
	return cfg, nil
}

// CreateSecondaryMaster creates a konfigadm config for a secondary master.
func CreateSecondaryMaster(platform *platform.Platform) (*konfigadm.Config, error) {
	cfg, err := baseKonfig(platform.Master.KonfigadmFile, platform)
	if err != nil {
		return nil, fmt.Errorf("createSecondaryMaster: failed to get baseKonfig: %v", err)
	}
	if err := addControlPlaneJoinConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("failed to add kubeadm config: %v", err)
	}

	files, err := kubeadm.GetFilesToMountForSecondary(platform)
	if err != nil {
		return nil, err
	}
	for file, content := range files {
		cfg.Files[file] = content
	}

	for _, cmd := range downloadCustomClusterSigningFiles {
		cfg.AddCommand(cmd)
	}
	cfg.AddCommand(kubeadmNodeJoinCmd)
	return cfg, nil
}

// CreateWorker creates a konfigadm config for a worker in node group nodegroup
func CreateWorker(nodegroup string, platform *platform.Platform) (*konfigadm.Config, error) {
	if nodegroup == "" {
		nodegroup = "default"
		if platform.Nodes == nil {
			platform.Nodes = make(map[string]types.VM)
		}

		if _, ok := platform.Nodes[nodegroup]; !ok {
			platform.Nodes[nodegroup] = types.VM{}
		}
	}
	if platform.Nodes == nil {
		return nil, fmt.Errorf("must specify 'nodes' in karina config")
	}
	if _, ok := platform.Nodes[nodegroup]; !ok {
		return nil, fmt.Errorf("node group %s not found", nodegroup)
	}

	node := platform.Nodes[nodegroup]
	baseConfig := node.KonfigadmFile
	cfg, err := baseKonfig(baseConfig, platform)
	if err != nil {
		return nil, fmt.Errorf("createWorker: failed to get baseKonfig: %v", err)
	}
	if err := addJoinKubeadmConfig(platform, cfg, node); err != nil {
		return nil, fmt.Errorf("failed to add kubeadm config: %v", err)
	}

	cfg.AddCommand(kubeadmNodeJoinCmd)
	return cfg, nil
}

// baseKonfig generates a base konfigadm configuration.
// It copies in the required environment variables and
// initial commands.
//
//nolint:unparam
func baseKonfig(initialKonfigadmFile string, platform *platform.Platform) (*konfigadm.Config, error) {
	var cfg *konfigadm.Config
	var err error
	if initialKonfigadmFile == "" {
		cfg, err = konfigadm.NewConfig().Build()
	} else {
		cfg, err = konfigadm.NewConfig(initialKonfigadmFile).Build()
	}
	if err != nil {
		return nil, fmt.Errorf("baseKonfig: failed to get config: %v", err)
	}

	for k, v := range envVars {
		cfg.Environment[k] = v
	}

	// update hosts file with hostname
	cfg.AddCommand(updateHostsFileCmd)
	return cfg, nil
}

// addInitKubeadmConfig derives the initial kubeadm config for a cluster from its platform
// config and adds it to its konfigadm files
func addInitKubeadmConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	if platform.Kubernetes.AuditConfig.PolicyFile != "" {
		// clusters audit policy files are injected into the machine via konfigadm
		ap := files.SafeRead(platform.Kubernetes.AuditConfig.PolicyFile)
		if ap == "" {
			return fmt.Errorf("unable to read audit policy file")
		}
		cfg.Files[kubeadm.AuditPolicyPath] = ap
	}

	cluster := kubeadm.NewClusterConfig(platform)
	data, err := yaml.Marshal(cluster)
	return addKubeadmConf(platform, data, err, cfg)
}

func addKubeadmConf(platform *platform.Platform, data []byte, err error, cfg *konfigadm.Config) error {
	if err != nil {
		return fmt.Errorf("addInitKubeadmConfig: failed to marshal cluster config: %v", err)
	}
	if platform.PlatformConfig.Trace {
		logger.Tracef("Using kubeadm config: \n%s", string(data))
	}
	cfg.Files["/etc/kubernetes/kubeadm.conf"] = string(data)
	return nil
}

// addJoinKubeadmConfig derives the initial kubeadm config for a cluster from its platform
// config and adds it to its konfigadm files
func addJoinKubeadmConfig(platform *platform.Platform, cfg *konfigadm.Config, node types.VM) error {
	data, err := kubeadm.NewJoinConfiguration(platform, node)
	return addKubeadmConf(platform, data, err, cfg)
}

// addJoinKubeadmConfig derives the initial kubeadm config for a cluster from its platform
// config and adds it to its konfigadm files
func addControlPlaneJoinConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	data, err := kubeadm.NewControlPlaneJoinConfiguration(platform)
	return addKubeadmConf(platform, data, err, cfg)
}
