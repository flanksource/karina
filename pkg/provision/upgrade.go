package provision

import (
	"fmt"
	"time"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
)

// Upgrade the kubernetes control plane to the declared version
func Upgrade(platform *platform.Platform) error {
	cluster, err := GetCluster(platform)
	if err != nil {
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
		out, err := platform.Executef(toUpgrade[0], 5*time.Minute, "kubeadm upgrade apply -y %s", platform.Kubernetes.Version)
		if err != nil {
			return fmt.Errorf("failed to upgrade: %s, %s", err, out)
		}
		platform.Infof("Completed upgrade via %s: %s", toUpgrade[0], out)
		toUpgrade = toUpgrade[1:]
	}
	for _, node := range toUpgrade {
		out, err := platform.Executef(node, 5*time.Minute, "kubeadm upgrade node")
		if err != nil {
			return fmt.Errorf("failed to upgrade: %s, %s", err, out)
		}
		platform.Infof("Upgraded node via %s: %s", node, out)
	}

	return nil
}
