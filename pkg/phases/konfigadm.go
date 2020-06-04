package phases

import (
	"fmt"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/logger"
	"gopkg.in/flanksource/yaml.v3"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
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

const updateHostsFileCmd = "echo $(ifconfig ens160 | grep inet | awk '{print $2}' | head -n1 ) $(hostname) >> /etc/hosts"
const kubeadmInitCmd = "kubeadm init --config /etc/kubernetes/kubeadm.conf -v 5 | tee /var/log/kubeadm.log"
const kubeadmNodeJoinCmd = "kubeadm join --config /etc/kubernetes/kubeadm.conf -v 5 | tee /var/log/kubeadm.log"

const noCAErrorText = `Must specify a 'ca'' section in the platform config.
e.g.:
ca:
   cert: .certs/root-ca-crt.pem
   privateKey: .certs/root-ca-key.pem
   password: foobar

CA certs are generated using karina ca generate
`

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
	if err := addAuditConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to add audit config: %v", err)
	}
	if err := addEncryptionConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to add encryption config: %v", err)
	}
	if err := addCerts(platform, cfg); err != nil {
		return nil, errors.Wrap(err, "failed to add certs")
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
	if err := addAuditConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("failed to add audit config: %v", err)
	}
	if err := addAuditConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to add audit config: %v", err)
	}
	if err := addEncryptionConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("createPrimaryMaster: failed to add encryption config: %v", err)
	}
	if err = addCerts(platform, cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to add certs")
	}
	cfg.AddCommand(kubeadmNodeJoinCmd)
	return cfg, nil
}

// CreateWorker creates a konfigadm config for a worker in node group nodegroup
func CreateWorker(nodegroup string, platform *platform.Platform) (*konfigadm.Config, error) {
	if platform.Nodes == nil {
		return nil, fmt.Errorf("CreateWorker failed to create worker - nil Nodes supplied")
	}
	if _, ok := platform.Nodes[nodegroup]; !ok {
		return nil, fmt.Errorf("CreateWorker failed to create worker - supplied nodegroup not found")
	}

	node := platform.Nodes[nodegroup]
	baseConfig := node.KonfigadmFile
	cfg, err := baseKonfig(baseConfig, platform)
	if err != nil {
		return nil, fmt.Errorf("createWorker: failed to get baseKonfig: %v", err)
	}
	if err := addJoinKubeadmConfig(platform, cfg); err != nil {
		return nil, fmt.Errorf("failed to add kubeadm config: %v", err)
	}

	cfg.AddCommand(kubeadmNodeJoinCmd)
	return cfg, nil
}

// baseKonfig generates a base konfigadm configuration.
// It copies in the required environment variables and
// initial commands.
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

// addAuditConfig derives the initial audit config for a cluster from its platform
// config and adds it to its konfigadm files
func addAuditConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	if pf := platform.Kubernetes.AuditConfig.PolicyFile; pf != "" {
		// clusters audit policy files are injected into the machine via konfigadm
		ap := files.SafeRead(pf)
		if ap == "" {
			return fmt.Errorf("unable to read audit policy file %v", pf)
		}
		cfg.Files[kubeadm.AuditPolicyPath] = ap
	}
	return nil
}

// addEncryptionConfig derives the initial encryption config for a cluster from its platform
// config and adds it to its konfigadm files
func addEncryptionConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	if ef := platform.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile; ef != "" {
		// clusters encryption provider files are injected into the machine via konfigadm
		ep := files.SafeRead(ef)
		if ep == "" {
			return fmt.Errorf("unable to encryption provider file %v", ef)
		}
		cfg.Files[kubeadm.EncryptionProviderConfigPath] = ep
	}
	return nil
}

// addCerts derives certs and key files for a cluster from its platform
// config and adds the cert and key files to its konfigadm files
func addCerts(platform *platform.Platform, cfg *konfigadm.Config) error {
	if platform.CA == nil {
		return errors.New(noCAErrorText)
	}

	clusterCA := certs.NewCertificateBuilder("kubernetes-ca").CA().Certificate
	platformCA, err := platform.GetCA()
	if err != nil {
		return fmt.Errorf("Error getting CA: %v", err)
	}
	clusterCA, err = platformCA.SignCertificate(clusterCA, 10)
	if err != nil {
		return fmt.Errorf("addCerts: failed to sign certificate: %v", err)
	}

	// plus any cert signed by this cluster specific CA
	crt := string(clusterCA.EncodedCertificate()) + "\n"
	// any cert signed by the global CA should be allowed
	platformCA, err = platform.GetCA()
	if err != nil {
		return fmt.Errorf("Error getting CA: %v", err)
	}
	crt = crt + string(platformCA.GetPublicChain()[0].EncodedCertificate()) + "\n"
	// csrsigning controller doesn't like having more than 1 CA cert passed to it
	cfg.Files["/etc/kubernetes/pki/csr-ca.crt"] = string(clusterCA.EncodedCertificate())
	cfg.Files["/etc/kubernetes/pki/csr-ca.key"] = string(clusterCA.EncodedPrivateKey())

	cfg.Files["/etc/kubernetes/pki/ca.crt"] = crt
	cfg.Files["/etc/kubernetes/pki/ca.key"] = string(clusterCA.EncodedPrivateKey())
	cfg.Files["/etc/ssl/certs/openid-ca.pem"] = string(platform.GetIngressCA().GetPublicChain()[0].EncodedCertificate())
	return nil
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
func addJoinKubeadmConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	data, err := kubeadm.NewJoinConfiguration(platform)
	return addKubeadmConf(platform, data, err, cfg)
}

// addJoinKubeadmConfig derives the initial kubeadm config for a cluster from its platform
// config and adds it to its konfigadm files
func addControlPlaneJoinConfig(platform *platform.Platform, cfg *konfigadm.Config) error {
	data, err := kubeadm.NewControlPlaneJoinConfiguration(platform)
	return addKubeadmConf(platform, data, err, cfg)
}
