package provision

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/k8s/etcd"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Status(p *platform.Platform) error {
	if err := WithVmwareCluster(p); err != nil {
		return err
	}

	vmList, err := p.Cluster.GetMachines()
	if err != nil {
		return fmt.Errorf("status: failed to get VMs: %v", err)
	}

	vms := make(map[string]map[string]string)
	for _, vm := range vmList {
		attributes, err := vm.GetAttributes()
		if err != nil {
			attributes["error"] = fmt.Sprintf("%s", err)
		}
		ip, err := vm.GetIP(1 * time.Second)
		if err != nil {
			attributes["ip"] = fmt.Sprintf("Error: %v", err)
		} else {
			attributes["ip"] = ip
		}
		vms[vm.Name()] = attributes
	}

	client, err := p.GetClientset()

	if err != nil {
		return fmt.Errorf("status: failed to get clientset: %v", err)
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("%+v", err)
		return err
	}
	cert, err := kubeadm.UploadEtcdCerts(p)
	if err != nil {
		p.Warnf("Could not find etcd ca certs: %v", err)
	}

	etcdClientGenerator, err := p.GetEtcdClientGenerator(cert)
	if err != nil {
		p.Warnf("Could not create etcd client generator: %v", err)
	}

	if err != nil {
		p.Warnf("Could not get etcdClient: %v", err)
	}
	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "NAME\tSTATUS\tETCD\tIP\tAGE\tTEMPLATE\tCPU\tMEM\tOS\tKERNEL\tCRI\t\n")
	for _, node := range list.Items {
		attributes := vms[node.Name]
		delete(vms, node.Name)
		fmt.Fprintf(w, "%s\t", node.Name)
		fmt.Fprintf(w, "%s\t", k8s.GetNodeStatus(node))
		if k8s.IsMasterNode(node) {
			fmt.Fprintf(w, "%s\t", getEtcdHealth(node, etcdClientGenerator))
		} else {
			fmt.Fprintf(w, "\t")
		}

		fmt.Fprintf(w, "%s\t", attributes["ip"])
		created, _ := time.Parse("02Jan06-15:04:05", attributes["CreatedDate"])
		age := time.Since(created).Round(time.Minute)
		fmt.Fprintf(w, "%s\t", age)
		fmt.Fprintf(w, "%s\t", attributes["Template"])
		fmt.Fprintf(w, "%d/%d\t", node.Status.Allocatable.Cpu().Value(), node.Status.Capacity.Cpu().Value())
		fmt.Fprintf(w, "%s/%s\t", gb(node.Status.Allocatable.Memory().Value()), gb(node.Status.Capacity.Memory().Value()))
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.OSImage)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.KernelVersion)
		fmt.Fprintf(w, "%s\t", node.Status.NodeInfo.ContainerRuntimeVersion)
		fmt.Fprintf(w, "\n")

	}
	_ = w.Flush()

	for vm := range vms {
		fmt.Printf("%s VM not in cluster\n", console.Redf(vm))
	}
	return nil
}

func getEtcdHealth(node v1.Node, etcdClientGenerator *etcd.EtcdClientGenerator) string {
	if etcdClientGenerator == nil {
		return "<nil>"
	}
	etcdClient, err := etcdClientGenerator.ForNode(context.Background(), node.Name)
	if err != nil {
		return fmt.Sprintf("Failed to get etcd client for %s: %v", node.Name, err)
	}
	s := ""

	status, err := etcdClient.EtcdClient.Status(context.Background(), etcdClient.EtcdClient.Endpoints()[0])
	if err != nil {
		s += fmt.Sprintf("cannot get status: %v", err)
	}
	s += fmt.Sprintf("v%s size: %s ", status.Version, size(status.DbSize))

	alarms, err := etcdClient.Alarms(context.Background())
	if err != nil {
		return fmt.Sprintf("Failed to get alarms for %s: %v", node.Name, err)
	}
	for _, alarm := range alarms {
		s += fmt.Sprintf("%v ", alarm.Type)
	}

	if etcdClient.LeaderID == etcdClient.MemberID {
		s += "Leader"
	}

	return s
}

func gb(bytes int64) string {
	return fmt.Sprintf("%d", int64(bytes/1024/1024/1024))
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
