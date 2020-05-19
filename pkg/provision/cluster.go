package provision

import (
	"context"
	"fmt"
	"sort"

	"github.com/moshloop/platform-cli/pkg/k8s"
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
	*platform.Platform
	Nodes      NodeMachines
	Orphans    []types.Machine
	Kubernetes kubernetes.Interface
	Etcd       *etcd.EtcdClientGenerator
}

func (cluster *Cluster) connectToEtcd() error {
	if cluster.Etcd != nil {
		return nil
	}

	// upload etcd certs so that we can connect to etcd
	cert, err := kubeadm.UploadEtcdCerts(cluster.Platform)
	if err != nil {
		return err
	}

	etcdClientGenerator, err := cluster.Platform.GetEtcdClientGenerator(cert)
	if err != nil {
		return err
	}
	cluster.Etcd = etcdClientGenerator
	return nil
}
func (cluster *Cluster) GetEtcdClient(node v1.Node) (*etcd.Client, error) {
	if err := cluster.connectToEtcd(); err != nil {
		return nil, err
	}
	return cluster.Etcd.ForNode(context.TODO(), node.Name)
}

func (cluster *Cluster) GetEtcdLeader() (*etcd.Client, error) {
	if err := cluster.connectToEtcd(); err != nil {
		return nil, err
	}
	list, err := cluster.Kubernetes.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return cluster.Etcd.ForLeader(context.TODO(), list)
}

func (n NodeMachines) Less(i, j int) bool {
	if n[j].Machine == nil || n[i].Machine == nil {
		return true
	}
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
		Kubernetes: client}
	cluster.Platform = platform
	return cluster, nil
}

func (cluster *Cluster) GetHealth(node v1.Node) string {
	if err := cluster.connectToEtcd(); err != nil {
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

func (cluster *Cluster) Cordon(node v1.Node) error {
	ctx := context.TODO()
	if k8s.IsMasterNode(node) {

		// we always interact via the etcd leader, as a previous node may have become unavailable
		leaderClient, err := cluster.GetEtcdLeader()
		if err != nil {
			return err
		}
		cluster.Infof("etcd leader is: %s", leaderClient.Name)

		members, err := leaderClient.Members(ctx)
		if err != nil {
			return err
		}
		var etcdMember *etcd.Member
		var etcdMembersNames []string
		var candidateLeader *etcd.Member
		for _, member := range members {
			etcdMembersNames = append(etcdMembersNames, member.Name)
			if member.Name == node.Name {
				// find the etcd member for the node
				etcdMember = member
			}
			if member.Name != leaderClient.Name {
				// choose a potential candidate to move the etcd leader
				candidateLeader = member
			}
		}
		if etcdMember == nil {
			cluster.Warnf("%s has already been removed from etcd cluster", node.Name)
		} else {
			if etcdMember.ID == leaderClient.MemberID {
				cluster.Infof("Moving etcd leader from %s to %s", node.Name, candidateLeader.Name)
				if err := leaderClient.MoveLeader(ctx, candidateLeader.ID); err != nil {
					return fmt.Errorf("failed to move leader: %v", err)
				}
			}

			cluster.Infof("Removing etcd member %s", node.Name)
			if err := leaderClient.RemoveMember(ctx, etcdMember.ID); err != nil {
				return err
			}
		}

		// proactively remove server from consul so that we can get a new connection to k8s
		if err := cluster.GetConsulClient().RemoveMember(node.Name); err != nil {
			return err
		}
		// reset the connection to the existing master (which may be the one we just removed)
		cluster.Platform.ResetMasterConnection()
		// wait for a new connection to be healthy before continuing
		if err := cluster.Platform.WaitFor(); err != nil {
			return err
		}
	}
	return cluster.Platform.Cordon(node.Name)
}
