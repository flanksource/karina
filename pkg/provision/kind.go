package provision

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/flanksource/yaml"
	"github.com/pkg/errors"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
)

var (
	kindCADir = "/etc/flanksource/ingress-ca"
)

// KindCluster provision or create a kubernetes cluster
func KindCluster(platform *platform.Platform) error {
	kubeadmPatches, err := createKubeAdmPatches(platform)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubeadm patches")
	}

	caPath, err := filepath.Abs(platform.IngressCA.Cert)
	if err != nil {
		return errors.Wrap(err, "failed to expand ca file path")
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
				ExtraMounts: []kindapi.Mount{
					{
						ContainerPath: kindCADir,
						HostPath:      path.Dir(caPath),
						Readonly:      true,
					},
				},
			},
		},
	}

	yml, err := yaml.Marshal(kindConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
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

func createKubeAdmPatches(platform *platform.Platform) ([]string, error) {
	clusterConfig := kubeadm.NewClusterConfig(platform)
	clusterConfig.ControlPlaneEndpoint = ""
	clusterConfig.ClusterName = ""
	clusterConfig.APIServer.CertSANs = nil
	clusterConfig.APIServer.ExtraVolumes = []api.HostPathMount{
		{
			Name:      "oidc-certificates",
			HostPath:  path.Join(kindCADir, filepath.Base(platform.IngressCA.Cert)),
			MountPath: "/etc/ssl/oidc/ingress-ca.pem",
			ReadOnly:  true,
			PathType:  api.HostPathFile,
		},
	}
	clusterConfig.APIServer.ExtraArgs["oidc-ca-file"] = "/etc/ssl/oidc/ingress-ca.pem"
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
