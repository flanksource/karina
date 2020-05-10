package provision

import (
	"context"
	"fmt"
	"sort"

	"github.com/moshloop/platform-cli/pkg/k8s/etcd"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NodeMachine struct {
	Node    v1.Node
	Machine types.Machine
}

type NodeMachines []NodeMachine

type Cluster struct {
	Nodes      NodeMachines
	Kubernetes kubernetes.Interface
	Etcd       *etcd.EtcdClientGenerator
}

func (cluster *Cluster) GetEtcdClient(node v1.Node) (*etcd.Client, error) {
	return cluster.Etcd.ForNode(context.TODO(), node.Name)
}

func (cluster *Cluster) GetEtcdLeader() (*etcd.Client, error) {
	list, err := cluster.Kubernetes.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return cluster.Etcd.ForLeader(context.TODO(), list)
}

func (n NodeMachines) Less(i, j int) bool {
	return n[j].Machine.GetAge() < n[i].Machine.GetAge()
}

func (n NodeMachines) Len() int {
	return len(n)
}

func (n NodeMachines) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func GetCluster(platform *platform.Platform) (*Cluster, error) {
	client, err := platform.GetClientset()
	if err != nil {
		return nil, err
	}
	if err := WithVmwareCluster(platform); err != nil {
		return nil, err
	}

	// make sure admin kubeconfig is available
	platform.GetKubeConfig() // nolint: errcheck
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}

	// upload control plane certs first
	if _, err := kubeadm.UploadControlPlaneCerts(platform); err != nil {
		return nil, err
	}

	// upload etcd certs so that we can connect to etcd
	cert, err := kubeadm.UploadEtcdCerts(platform)
	if err != nil {
		return nil, err
	}

	etcdClientGenerator, err := platform.GetEtcdClientGenerator(cert)
	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := NodeMachines{}
	for _, node := range list.Items {
		machine, err := platform.Cluster.GetMachine(node.Name)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, NodeMachine{Node: node, Machine: machine})
	}
	// roll nodes from oldest to newest
	sort.Sort(nodes)
	return &Cluster{Nodes: nodes, Etcd: etcdClientGenerator, Kubernetes: client}, nil
}

func (cluster *Cluster) GetHealth(node v1.Node) string {
	if cluster.Etcd == nil {
		return "<nil>"
	}
	etcdClient, err := cluster.Etcd.ForNode(context.Background(), node.Name)
	if err != nil {
		return fmt.Sprintf("Failed to get etcd client for %s: %v", node.Name, err)
	}
	s := ""

	status, err := etcdClient.EtcdClient.Status(context.Background(), etcdClient.EtcdClient.Endpoints()[0])
	if err != nil {
		s += fmt.Sprintf("cannot get status: %v", err)
	}
	s += fmt.Sprintf("v%s (%s) ", status.Version, size(status.DbSize))

	alarms, err := etcdClient.Alarms(context.Background())
	if err != nil {
		return fmt.Sprintf("Failed to get alarms for %s: %v", node.Name, err)
	}
	for _, alarm := range alarms {
		s += fmt.Sprintf("%v ", alarm.Type)
	}

	if etcdClient.LeaderID == etcdClient.MemberID {
		s += "**"
	}
	return s
}
