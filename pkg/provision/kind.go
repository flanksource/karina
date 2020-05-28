package provision

import (
	"fmt"

	"github.com/flanksource/commons/files"

	"github.com/pkg/errors"

	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/flanksource/yaml.v3"

	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
)

var (
	kindCADir = "/etc/flanksource/ingress-ca"
)

// KindCluster provisions a new Kind cluster
func KindCluster(platform *platform.Platform) error {
	kubeadmPatches, err := createKubeAdmPatches(platform)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubeadm patches")
	}

	var extraMounts []kindapi.Mount
	if platform.IngressCA != nil {
		caPath, err := filepath.Abs(platform.IngressCA.Cert)
		if err != nil {
			return errors.Wrap(err, "failed to expand ca file path")
		}
		extraMounts = []kindapi.Mount{
			{
				ContainerPath: kindCADir,
				HostPath:      path.Dir(caPath),
				Readonly:      true,
			}}
	}

	kindConfig := kindapi.Cluster{
		TypeMeta: kindapi.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Networking: kindapi.Networking{
			DisableDefaultCNI: true,
		},
		Nodes: []kindapi.Node{
			{
				Role:  "control-plane",
				Image: fmt.Sprintf("kindest/node:%s", platform.Kubernetes.Version),
				ExtraPortMappings: []kindapi.PortMapping{
					{
						ContainerPort: 80,
						HostPort:      80,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
					{
						ContainerPort: 443,
						HostPort:      443,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
					{
						ContainerPort: 6443,
						HostPort:      6443,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
				},
				KubeadmConfigPatches: kubeadmPatches,
				ExtraMounts:          extraMounts,
			},
		},
	}

	err = configureAuditMappings(platform, &kindConfig)
	if err != nil {
		return err
	}

	err = configureEncryptionMappings(platform, &kindConfig)
	if err != nil {
		return err
	}

	yml, err := yaml.Marshal(kindConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	if platform.PlatformConfig.Trace {
		platform.Infof(string(yml))
	}

	tmpfile, err := ioutil.TempFile("", "kind.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to create tempfile")
	}
	defer os.Remove(tmpfile.Name())

	if err := ioutil.WriteFile(tmpfile.Name(), yml, 0644); err != nil {
		return errors.Wrap(err, "failed to write kind config file")
	}

	if platform.DryRun {
		fmt.Println(string(yml))
		return nil
	}

	kind := platform.GetBinary("kind")
	kubeConfig, err := platform.GetKubeConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kube config")
	}

	if err := kind("create cluster --config %s --kubeconfig %s", tmpfile.Name(), kubeConfig); err != nil {
		return err
	}

	// delete the default storageclass created by kind as we install our own
	return platform.GetKubectl()("delete sc standard")
}

// createKubeAdmPatches reads a Platform config, creates a new ClusterConfiguration from it and
// then extracts a slice of kind-specific KubeAdm patches from it.
func createKubeAdmPatches(platform *platform.Platform) ([]string, error) {
	clusterConfig := kubeadm.NewClusterConfig(platform)
	clusterConfig.ControlPlaneEndpoint = ""
	clusterConfig.ClusterName = platform.Name
	clusterConfig.APIServer.CertSANs = nil

	vol := &clusterConfig.APIServer.ExtraVolumes

	if platform.IngressCA != nil {
		*vol = append(*vol, api.HostPathMount{
			Name:      "oidc-certificates",
			HostPath:  path.Join(kindCADir, filepath.Base(platform.IngressCA.Cert)),
			MountPath: "/etc/ssl/oidc/ingress-ca.pem",
			ReadOnly:  true,
			PathType:  api.HostPathFile,
		})

		clusterConfig.APIServer.ExtraArgs["oidc-ca-file"] = "/etc/ssl/oidc/ingress-ca.pem"
	}
	clusterConfig.ControllerManager.ExtraArgs = nil
	clusterConfig.CertificatesDir = ""
	clusterConfig.Networking.PodSubnet = ""
	clusterConfig.Networking.ServiceSubnet = ""

	kubeadmPatches := []interface{}{
		clusterConfig,
		kubeadm.NewInitConfig(platform),
	}

	result := make([]string, len(kubeadmPatches))
	for i, x := range kubeadmPatches {
		yml, err := yaml.Marshal(x)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode yaml for kubeadm patch %v", x)
		}
		result[i] = string(yml)
	}

	return result, nil
}

// configureAuditMappings configures the mapping of the audit policy config files into the
// KIND cluster config
func configureAuditMappings(platform *platform.Platform, kindConfig *kindapi.Cluster) error {
	policyFile := platform.Kubernetes.AuditConfig.PolicyFile
	if policyFile == "" {
		platform.Debugf("No audit policy specified for KIND cluster")
		return nil
	}
	// for kind clusters audit policy files are mapped in via a dual
	// host -> master,
	// master -> kube-api-server pod
	// mapping

	absFile, err := filepath.Abs(policyFile)
	if err != nil {
		return errors.Wrap(err, "failed to expand audit policy file path")
	}
	if content := files.SafeRead(absFile); len(content) == 0 {
		return fmt.Errorf("failed to read audit policy file %v", absFile)
	}

	mnts := &kindConfig.Nodes[0].ExtraMounts

	*mnts = append(*mnts, kindapi.Mount{
		ContainerPath: kubeadm.AuditPolicyPath,
		HostPath:      absFile,
		Readonly:      true,
	})

	return nil
}

// configureEncryptionMappings configures the mapping of the encryption provider config files into the
// KIND cluster config
func configureEncryptionMappings(platform *platform.Platform, kindConfig *kindapi.Cluster) error {
	encryptionFile := platform.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile
	if encryptionFile == "" {
		platform.Debugf("No encryption provider config specified for KIND cluster")
		return nil
	}
	// for kind clusters encryption provider config files are mapped in via a dual
	// host -> master,
	// master -> kube-api-server pod
	// mapping

	absFile, err := filepath.Abs(encryptionFile)
	if err != nil {
		return errors.Wrapf(err, "failed to expand encryption provider file path for %v", encryptionFile)
	}
	if content := files.SafeRead(absFile); len(content) == 0 {
		return fmt.Errorf("failed to read encryption provider file %v", absFile)
	}

	mnts := &kindConfig.Nodes[0].ExtraMounts

	*mnts = append(*mnts, kindapi.Mount{
		ContainerPath: kubeadm.EncryptionProviderConfigPath,
		HostPath:      absFile,
		Readonly:      true,
	})
	return nil
}
