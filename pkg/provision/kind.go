package provision

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/exec"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	"gopkg.in/flanksource/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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

	files, err := kubeadm.GetFilesToMountForPrimary(p)
	if err != nil {
		return err
	}

	cacheDir := os.ExpandEnv("$CACHE_DIR")
	if cacheDir == "" {
		cacheDir = os.ExpandEnv("$HOME")
	}

	for file, content := range files {
		cache := fmt.Sprintf("%s/.kind/%s%s", cacheDir, p.Name, file)
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
		hostPort, err := strconv.Atoi(from)
		if err != nil {
			return errors.Wrapf(err, "failed to convert port %s to int", from)
		}
		portMappings = append(portMappings, kindapi.PortMapping{
			ContainerPort: to,
			HostPort:      int32(hostPort),
			Protocol:      kindapi.PortMappingProtocolTCP,
		})
	}
	if p.Kind.Image == "" {
		p.Kind.Image = fmt.Sprintf("kindest/node:%s", p.Kubernetes.Version)
	} else if !strings.Contains(p.Kind.Image, ":") {
		p.Kind.Image = fmt.Sprintf("%s:%s", p.Kind.Image, p.Kubernetes.Version)
	}

	nodes := []kindapi.Node{
		{
			Role:                 "control-plane",
			Image:                p.Kind.Image,
			ExtraPortMappings:    portMappings,
			KubeadmConfigPatches: kubeadmPatches,
			ExtraMounts:          extraMounts,
		},
	}

	if p.Kind.WorkerCount > 0 {
		workerNodes := kindWorkerNodesConfig(p, nodes[0])
		nodes = append(nodes, workerNodes...)
	}

	kindConfig := kindapi.Cluster{
		TypeMeta: kindapi.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Networking: kindapi.Networking{
			DisableDefaultCNI: true,
			PodSubnet:         p.PodSubnet,
		},
		Nodes: nodes,
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

	debugString := ""
	if p.Logger.IsDebugEnabled() {
		debugString = " -v 1"
	}
	if p.Logger.IsTraceEnabled() || p.PlatformConfig.Trace {
		debugString = " -v 6"
	}

	kind := p.GetBinary("kind")

	kindExecution := fmt.Sprintf("create cluster --config %s --kubeconfig %s --name %s%s", tmpfile.Name(), p.KubeConfigPath, p.Name, debugString)
	if p.PlatformConfig.Trace {
		p.Infof("Running Kind with: %s", kindExecution)
	}

	if err := kind(kindExecution); err != nil {
		return err
	}

	client, err := p.GetClientset()
	if err != nil {
		return err
	}

	// delete the default storageclass created by kind as we install our own
	if err := client.StorageV1().StorageClasses().Delete(context.TODO(), "standard", metav1.DeleteOptions{}); err != nil {
		return err
	}

	if err := createEtcdCertificateSecret(p); err != nil {
		return errors.Wrap(err, "failed to create etcd-certs secret")
	}
	return nil
}

func createEtcdCertificateSecret(p *platform.Platform) error {
	podName := fmt.Sprintf("etcd-%s-control-plane", p.Name)
	if err := p.WaitForPod("kube-system", podName, 3*time.Minute, v1.PodRunning); err != nil {
		return errors.Wrapf(err, "failed to wait for pod %s to be running", podName)
	}

	nodeName := fmt.Sprintf("%s-control-plane", p.Name)
	caCrtFile, err := ioutil.TempFile("", "ca.crt")
	if err != nil {
		return errors.Wrap(err, "failed to create ca.crt temp file")
	}
	defer os.Remove(caCrtFile.Name())
	caKeyFile, err := ioutil.TempFile("", "ca.key")
	if err != nil {
		return errors.Wrap(err, "failed to create ca.key temp file")
	}
	defer os.Remove(caKeyFile.Name())

	cmd := fmt.Sprintf("docker cp %s:/etc/kubernetes/pki/etcd/ca.crt %s", nodeName, caCrtFile.Name())
	if err := exec.Execf(cmd); err != nil {
		return errors.Wrap(err, "failed to copy ca.crt")
	}
	cmd = fmt.Sprintf("docker cp %s:/etc/kubernetes/pki/etcd/ca.key %s", nodeName, caKeyFile.Name())
	if err := exec.Execf(cmd); err != nil {
		return errors.Wrap(err, "failed to copy ca.key")
	}

	caCrt, err := ioutil.ReadFile(caCrtFile.Name())
	if err != nil {
		return errors.Wrap(err, "failed to read ca.crt file")
	}
	caKey, err := ioutil.ReadFile(caKeyFile.Name())
	if err != nil {
		return errors.Wrap(err, "failed to read ca.key file")
	}

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-certs",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			"tls.crt": caCrt,
			"tls.key": caKey,
		},
	}

	clientset, err := p.GetClientset()
	if err != nil {
		return errors.Wrap(err, "failed to get clientset")
	}
	if _, err := clientset.CoreV1().Secrets(metav1.NamespaceSystem).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to create secret %s in namespace %s", secret.Name, secret.Namespace)
	}
	return nil
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

func kindWorkerNodesConfig(p *platform.Platform, master kindapi.Node) []kindapi.Node {
	nodes := make([]kindapi.Node, p.Kind.WorkerCount)

	for i := 0; i < p.Kind.WorkerCount; i++ {
		nodes[i] = kindapi.Node{
			Role:                 kindapi.WorkerRole,
			Image:                master.Image,
			ExtraMounts:          master.ExtraMounts,
			KubeadmConfigPatches: master.KubeadmConfigPatches,
		}
	}

	return nodes
}
