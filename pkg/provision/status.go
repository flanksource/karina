package provision

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Status(p *platform.Platform) error {
	cluster, err := GetCluster(p)
	if err != nil {
		return err
	}

	if version, err := kubeadm.GetClusterVersion(p); err != nil {
		fmt.Printf("Cluster Version: %s\n", console.Redf("%s", err))
	} else {
		fmt.Printf("Cluster Version: %s\n", version)
	}

	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "NAME\tSTATUS\tAPI\tETCD\tIP\tAGE\tTEMPLATE\tCPU\tMEM\tOS\tKERNEL\tCRI\t\n")
	for _, nodeMachine := range cluster.Nodes {
		node := nodeMachine.Node
		fmt.Fprintf(w, "%s\t", node.Name)
		fmt.Fprintf(w, "%s\t", k8s.GetNodeStatus(node))
		if k8s.IsMasterNode(node) {
			fmt.Fprintf(w, "%s\t", kubeadm.GetNodeVersion(p, nodeMachine.Node))
			fmt.Fprintf(w, "%s\t", cluster.GetHealth(node))
		} else {
			fmt.Fprintf(w, "\t\t")
		}

		fmt.Fprintf(w, "%s\t", nodeMachine.Machine.IP())
		fmt.Fprintf(w, "%s\t", age(nodeMachine.Machine.GetAge()))
		fmt.Fprintf(w, "%s\t", nodeMachine.Machine.GetTemplate())
		fmt.Fprintf(w, "%d/%d\t", node.Status.Allocatable.Cpu().Value(), node.Status.Capacity.Cpu().Value())
		fmt.Fprintf(w, "%s/%s\t", gb(node.Status.Allocatable.Memory().Value()), gb(node.Status.Capacity.Memory().Value()))
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.OSImage)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.KernelVersion)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.ContainerRuntimeVersion)
		fmt.Fprintf(w, "\n")
	}
	_ = w.Flush()
	return nil
}

func age(t time.Duration) string {
	if t.Hours() > 24 {
		return fmt.Sprintf("%.0fd", t.Hours()/24)
	} else if t.Hours() > 1 {
		return fmt.Sprintf("%.0fh", t.Hours())
	} else {
		return fmt.Sprintf("%.0fm", t.Minutes())
	}
}

func gb(bytes int64) string {
	return fmt.Sprintf("%d", bytes/1024/1024/1024)
}

func size(bytes int64) string {
	if bytes > 1024*1024*1024 {
		return fmt.Sprintf("%.00d gb", bytes/1024/1024/1024)
	}
	if bytes > 1024*1024 {
		return fmt.Sprintf("%.00d mb", bytes/1024/1024)
	}
	if bytes > 1024 {
		return fmt.Sprintf("%.00d kb", bytes/1024)
	}
	return fmt.Sprintf("%.00d bytes", bytes)
}
