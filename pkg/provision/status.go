package provision

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PodStatus(p *platform.Platform, period time.Duration) error {
	client, err := p.GetClientset()
	if err != nil {
		return err
	}

	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "NAMESPACE\tNODE\tNAME\tPHASE\tREADY\tIP\tAGE\tRESTARTED\t \n")
	for _, pod := range pods.Items {
		lastRestarted := kommons.GetLastRestartTime(pod)
		if kommons.IsPodHealthy(pod) && (lastRestarted == nil || time.Since(*lastRestarted) > period) {
			continue
		}
		events, _ := p.GetEventsFor("Pod", &pod)
		var lastEvent *v1.Event
		if len(events) > 0 {
			lastEvent = &events[len(events)-1]
		}
		fmt.Fprintf(w, "%s\t", pod.Namespace)
		fmt.Fprintf(w, "%s\t", pod.Spec.NodeName)
		fmt.Fprintf(w, "%s\t", pod.Name)
		podStatus := k8s.GetPodStatus(pod)

		fmt.Fprintf(w, "%s\t", podStatus)
		if k8s.IsPodReady(pod) {
			fmt.Fprintf(w, "TRUE\t")
		} else {
			fmt.Fprintf(w, "FALSE\t")
		}
		fmt.Fprintf(w, "%s\t", pod.Status.PodIP)

		if pod.Status.StartTime != nil {
			fmt.Fprintf(w, "%s\t", age(time.Since(pod.Status.StartTime.Time)))
		} else {
			fmt.Fprintf(w, "\t")
		}

		if lastRestarted == nil {
			fmt.Fprintf(w, "\t")
		} else {
			fmt.Fprintf(w, "%s\t", age(time.Since(*lastRestarted)))
		}

		// ignore Started and Backoff events as they do not provide any diagnostic value
		if lastEvent != nil && lastEvent.Reason != "Started" && lastEvent.Reason != "BackOff" {
			fmt.Fprintf(w, "%s: %s", lastEvent.Reason, lastEvent.Message)
		}
		fmt.Fprint(w, kommons.GetContainerStatus(pod))
		fmt.Fprintf(w, "\t\n")
	}

	_ = w.Flush()
	return nil
}

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
		if kommons.IsMasterNode(node) {
			fmt.Fprintf(w, "%s\t", kubeadm.GetNodeVersion(p, nodeMachine.Node))
			fmt.Fprintf(w, "%s\t", cluster.GetHealth(node))
		} else {
			fmt.Fprintf(w, "\t\t")
		}

		ip, _ := nodeMachine.Machine.GetIP(5 * time.Second)
		fmt.Fprintf(w, "%s\t", ip)
		fmt.Fprintf(w, "%s\t", age(nodeMachine.Machine.GetAge()))
		fmt.Fprintf(w, "%s\t", nodeMachine.Machine.GetTemplate())
		fmt.Fprintf(w, "%d/%d\t", node.Status.Allocatable.Cpu().Value(), node.Status.Capacity.Cpu().Value())
		fmt.Fprintf(w, "%s/%s\t", gb(node.Status.Allocatable.Memory().Value()), gb(node.Status.Capacity.Memory().Value()))
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.OSImage)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.KernelVersion)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.ContainerRuntimeVersion)
		fmt.Fprintf(w, "\n")
	}

	for _, orphan := range cluster.Orphans {
		fmt.Fprintf(w, "%s\t", orphan.Name())
		fmt.Fprintf(w, "%s\t", "orphan")
		fmt.Fprintf(w, "\t\t")
		ip, _ := orphan.GetIP(5 * time.Second)
		fmt.Fprintf(w, "%s\t", ip)
		fmt.Fprintf(w, "%s\t", age(orphan.GetAge()))
		fmt.Fprintf(w, "%s\t", orphan.GetTemplate())
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
