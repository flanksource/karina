package provision

import (
	"fmt"
	"sort"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NodeMachine struct {
	Node    v1.Node
	Machine types.Machine
}

func (node NodeMachine) String() string {
	return fmt.Sprintf("%s (%s)", node.Node.Name, node.Machine.IP())
}

type NodeMachines []NodeMachine

func (n NodeMachines) Less(i, j int) bool {
	if n[j].Machine == nil || n[i].Machine == nil {
		return true
	}
	return n[i].Machine.GetAge() < n[j].Machine.GetAge()
}

func (n NodeMachines) Len() int {
	return len(n)
}

func (n NodeMachines) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n *NodeMachines) Push(x NodeMachine) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*n = append(*n, x)
}

func (n *NodeMachines) Pop() NodeMachine {
	old := *n
	l := len(old)
	x := old[l-1]
	*n = old[0 : l-1]
	return x
}

func (n *NodeMachines) PopN(count int) *[]NodeMachine {
	items := []NodeMachine{}

	for i := 0; i < count; {
		if n.Len() == 0 || len(items) == count {
			return &items
		}
		items = append(items, n.Pop())
	}
	return &items
}

type Cluster struct {
	*platform.Platform
	Nodes      NodeMachines
	Orphans    []types.Machine
	Kubernetes kubernetes.Interface
	Etcd       *EtcdClient
}

func GetCluster(platform *platform.Platform) (*Cluster, error) {
	if err := WithVmwareCluster(platform); err != nil {
		return nil, err
	}

	client, err := platform.GetClientset()
	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	machines, err := platform.Cluster.GetMachines()
	if err != nil {
		return nil, err
	}

	orphans := []types.Machine{}
	nodes := NodeMachines{}
	joinedNodes := map[string]bool{}
	for _, node := range list.Items {
		machine := machines[node.Name]
		joinedNodes[node.Name] = true
		if machine == nil {
			machine = types.NullMachine{Hostname: node.Name}
		}
		nodes = append(nodes, NodeMachine{Node: node, Machine: machine})
	}

	for name, machine := range machines {
		if joinedNodes[name] {
			continue
		}
		orphans = append(orphans, machine)
	}

	// roll nodes from oldest to newest
	sort.Sort(nodes)
	cluster := &Cluster{
		Nodes:      nodes,
		Orphans:    orphans,
		Kubernetes: client,
		Etcd:       GetEtcdClient(platform, client, ""),
	}
	cluster.Platform = platform
	return cluster, nil
}

func (cluster *Cluster) Terminate(node types.Machine) error {
	terminate(cluster.Platform, cluster.Etcd, node)
	return nil
}

func (cluster *Cluster) GetHealth(node v1.Node) string {
	return cluster.Etcd.GetHealth(node)
}

func (cluster *Cluster) Cordon(node v1.Node) error {
	return cluster.Platform.Cordon(node.Name)
}
