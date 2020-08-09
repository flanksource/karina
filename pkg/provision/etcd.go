package provision

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/k8s/etcd"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/platform"
	"go.etcd.io/etcd/clientv3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type EtcdClient struct {
	*platform.Platform
	Kubernetes        kubernetes.Interface
	PreferredHostname string
	Etcd              *etcd.EtcdClientGenerator
}

func GetEtcdClient(platform *platform.Platform, client kubernetes.Interface, PreferredHostname string) *EtcdClient {
	return &EtcdClient{Platform: platform, Kubernetes: client, PreferredHostname: PreferredHostname}
}

func (cluster *EtcdClient) connectToEtcd() error {
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

func (cluster *EtcdClient) getMemberStatus(name string) (*clientv3.StatusResponse, error) {
	if err := cluster.connectToEtcd(); err != nil {
		return nil, err
	}
	etcdClient, err := cluster.Etcd.ForNode(context.Background(), name)
	if err != nil {
		return nil, err
	}
	return etcdClient.EtcdClient.Status(context.Background(), etcdClient.EtcdClient.Endpoints()[0])
}

func (cluster *EtcdClient) GetHealth(node v1.Node) string {
	status, err := cluster.getMemberStatus(node.Name)
	if err != nil {
		return err.Error()
	}
	s := fmt.Sprintf("v%s (%s) ", status.Version, size(status.DbSize))
	return s
}

func (cluster *EtcdClient) RemoveMember(name string) error {
	client, err := cluster.GetEtcdLeader()
	if err != nil {
		return err
	}
	members, err := client.Members(context.TODO())
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.Name == name {
			return client.RemoveMember(context.TODO(), member.ID)
		}
	}
	return fmt.Errorf("member not found: %s", name)
}

func (cluster *EtcdClient) MoveLeader(name string) error {
	client, err := cluster.GetEtcdLeader()
	if err != nil {
		return err
	}
	members, err := client.Members(context.TODO())
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.Name == name {
			return client.MoveLeader(context.TODO(), member.ID)
		}
	}

	return fmt.Errorf("member not found: %s", name)
}

func (cluster *EtcdClient) GetEtcdClient(node v1.Node) (*etcd.Client, error) {
	if err := cluster.connectToEtcd(); err != nil {
		return nil, err
	}
	return cluster.Etcd.ForNode(context.TODO(), node.Name)
}

func (cluster *EtcdClient) GetOrphans() ([]string, error) {
	var orphans []string
	return orphans, nil
}

func (cluster *EtcdClient) GetEtcdLeader() (*etcd.Client, error) {
	if err := cluster.connectToEtcd(); err != nil {
		return nil, err
	}

	if cluster.PreferredHostname != "" {
		return cluster.Etcd.ForNode(context.TODO(), cluster.PreferredHostname)
	}
	list, err := cluster.Kubernetes.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return cluster.Etcd.ForLeader(context.TODO(), list)
}

func (cluster *EtcdClient) PrintStatus() error {
	if err := cluster.connectToEtcd(); err != nil {
		return fmt.Errorf("failed to connect to etcd: %v", err)
	}
	client, err := cluster.GetEtcdLeader()
	if err != nil {
		return fmt.Errorf("failed to connect to etcd: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "ID\tNAME\tK8S\tSTATUS\tVERSION\tSIZE\tALARMS \n")
	members, err := client.Members(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get members: %v", err)
	}
	for _, member := range members {
		fmt.Fprintf(w, "%d\t", member.ID)
		fmt.Fprintf(w, "%s\t", member.Name)
		node, err := cluster.Kubernetes.CoreV1().Nodes().Get(member.Name, metav1.GetOptions{})
		if err == nil {
			fmt.Fprintf(w, "%s\t", k8s.GetNodeStatus(*node))
		} else {
			fmt.Fprintf(w, "MISSING\t")
		}
		s := ""
		if member.ID == client.LeaderID {
			s = "LEADER"
		} else if member.IsLearner {
			s = "LEARNER"
		}
		status, err := cluster.getMemberStatus(member.Name)
		if err != nil {
			s += " " + err.Error()
			fmt.Fprintf(w, "%s\t\t\t", s)
		} else if status != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\t", s, status.Version, size(status.DbSizeInUse))
		}

		fmt.Fprintf(w, "%s\t", collections.ToString(member.Alarms))
		fmt.Fprintf(w, "\n")
	}
	_ = w.Flush()
	return nil
}
