package provision

import (
	"fmt"
	"time"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	"gopkg.in/yaml.v2"
)

const ClusterConfiguration = "ClusterConfiguration"

const upgradePrepCommand =
// install the correct version kubeadm
"apt-get install -y --allow-change-held-packages kubeadm=%s-00 ;" +
	// prevent it from being automatically updated
	" apt-mark hold kubeadm &&" +
	// download the most recent kubeadm configuration
	" (kubectl --kubeconfig /etc/kubernetes/admin.conf get cm kubeadm-config -o json -n kube-system | jq -r '.data.ClusterConfiguration' > /etc/kubernetes/kubeadm.conf)"
	// perform the upgrade
const upgradeCluster = "kubeadm upgrade apply -y --allow-experimental-upgrades --allow-release-candidate-upgrades --config /etc/kubernetes/kubeadm.conf %s"

const upgradeNode = "apt-get install -y --allow-change-held-packages kubeadm=%s-00;" +
	" apt-mark hold kubeadm &&" +
	" kubeadm upgrade node"

// Upgrade the kubernetes control plane to the declared version
func Upgrade(platform *platform.Platform) error {
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}

	newConfig := kubeadm.NewClusterConfig(platform)
	newData, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}
	kubeadmConfig := (*platform.GetConfigMap("kube-system", "kubeadm-config"))
	kubeadmConfig[ClusterConfiguration] = string(newData)

	// update kubeadm-config with any changes introduced since last provision/update
	if err := platform.CreateOrUpdateConfigMap("kubeadm-config", "kube-system", kubeadmConfig); err != nil {
		return err
	}

	version, err := kubeadm.GetClusterVersion(platform)
	if err != nil {
		return err
	}

	if platform.Kubernetes.Version == version {
		platform.Infof("At least 1 master has been upgraded to : %s\n", platform.Kubernetes.Version)
	} else {
		platform.Infof("Starting upgrade from %s to %s", version, platform.Kubernetes.Version)
	}

	var toUpgrade = []string{}
	var upgraded = []string{}

	for _, nodeMachine := range cluster.Nodes {
		node := nodeMachine.Node
		if !k8s.IsMasterNode(node) {
			continue
		}
		nodeVersion := kubeadm.GetNodeVersion(platform, node)
		if nodeVersion != platform.Kubernetes.Version {
			toUpgrade = append(toUpgrade, node.Name)
		} else {
			upgraded = append(upgraded, node.Name)
		}
	}

	platform.Infof("Nodes needing upgrade: %s", toUpgrade)
	platform.Infof("Nodes already upgraded: %s", upgraded)

	if len(upgraded) == 0 {
		out, err := platform.Executef(toUpgrade[0], 5*time.Minute, upgradePrepCommand, platform.Kubernetes.Version[1:])
		if err != nil {
			return fmt.Errorf("failed to prep for upgrade: %s, %s", err, out)
		}
		out, err = platform.Executef(toUpgrade[0], 5*time.Minute, upgradeCluster, platform.Kubernetes.Version)
		if err != nil {
			return fmt.Errorf("failed to upgrade: %s, %s", err, out)
		}
		platform.Infof("Completed upgrade via %s: %s", toUpgrade[0], out)
		toUpgrade = toUpgrade[1:]
	}
	for _, node := range toUpgrade {
		out, err := platform.Executef(node, 5*time.Minute, upgradeNode, platform.Kubernetes.Version[1:])
		if err != nil {
			return fmt.Errorf("failed to upgrade: %s, %s", err, out)
		}
		platform.Infof("Upgraded node via %s: %s", node, out)
	}

	return nil
}
