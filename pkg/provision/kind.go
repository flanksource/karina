package provision

import (
	"fmt"

	"github.com/pkg/errors"

	"io/ioutil"
	"os"
	"path"

	"gopkg.in/flanksource/yaml.v3"

	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
)

// KindCluster provisions a new Kind cluster
func KindCluster(p *platform.Platform) error {
	p.MasterDiscovery = platform.KindProvider{}
	p.ProvisionHook = platform.CompositeHook{}
	if p.Kubernetes.Version == "" {
		return fmt.Errorf("must specify kubernetes.version")
	}
	kubeadmPatches, err := createKubeAdmPatches(p)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubeadm patches")
	}

	var extraMounts []kindapi.Mount

	files, err := kubeadm.GetFilesToMount(p)
	if err != nil {
		return err
	}

	for file, content := range files {
		cache := fmt.Sprintf("%s/.kind/%s%s", os.ExpandEnv("$HOME"), p.Name, file)
		_ = os.MkdirAll(path.Dir(cache), 0755)
		if err := ioutil.WriteFile(cache, []byte(content), 0644); err != nil {
			return err
		}
		extraMounts = append(extraMounts, kindapi.Mount{
			ContainerPath: file,
			HostPath:      cache,
			Readonly:      true,
		})
	}

	portMappings := []kindapi.PortMapping{
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
	}

	for from, to := range p.Kind.PortMappings {
		portMappings = append(portMappings, kindapi.PortMapping{
			ContainerPort: to,
			HostPort:      from,
			Protocol:      kindapi.PortMappingProtocolTCP,
		})
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
				Role:                 "control-plane",
				Image:                fmt.Sprintf("kindest/node:%s", p.Kubernetes.Version),
				ExtraPortMappings:    portMappings,
				KubeadmConfigPatches: kubeadmPatches,
				ExtraMounts:          extraMounts,
			},
		},
	}

	yml, err := yaml.Marshal(kindConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	if p.PlatformConfig.Trace {
		p.Infof(string(yml))
	}

	tmpfile, err := ioutil.TempFile("", "kind.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to create tempfile")
	}
	defer os.Remove(tmpfile.Name())

	if err := ioutil.WriteFile(tmpfile.Name(), yml, 0644); err != nil {
		return errors.Wrap(err, "failed to write kind config file")
	}

	if p.DryRun {
		fmt.Println(string(yml))
		return nil
	}

	kind := p.GetBinary("kind")

	if err := kind("create cluster --config %s --kubeconfig %s --name %s", tmpfile.Name(), p.KubeConfigPath, p.Name); err != nil {
		return err
	}

	client, err := p.GetClientset()
	if err != nil {
		return err
	}

	// delete the default storageclass created by kind as we install our own
	return client.StorageV1().StorageClasses().Delete("standard", nil)
}

// createKubeAdmPatches reads a Platform config, creates a new ClusterConfiguration from it and
// then extracts a slice of kind-specific KubeAdm patches from it.
func createKubeAdmPatches(platform *platform.Platform) ([]string, error) {
	clusterConfig := kubeadm.NewClusterConfig(platform)
	clusterConfig.ControlPlaneEndpoint = ""
	clusterConfig.ClusterName = platform.Name
	clusterConfig.APIServer.CertSANs = nil
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
