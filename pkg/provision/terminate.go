package provision

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/karina/pkg/k8s/etcd"
	"github.com/flanksource/karina/pkg/platform"
)

// TerminateOrphans deletes all vm's that have not yet joined the cluster
func TerminateOrphans(platform *platform.Platform) error {
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}

	for _, orphan := range cluster.Orphans {
		time.Sleep(1 * time.Second) // sleep to allow for cancellation
		platform.Infof("Deleting %s", orphan.Name())
		cluster.Terminate(orphan)
	}
	return nil
}

// TerminateNodes deletes all of the specified nodes stops and deletes all VM's for a cluster;
func TerminateNodes(platform *platform.Platform, nodes []string) error {
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}
	toDelete := map[string]bool{}
	for _, node := range nodes {
		toDelete[node] = true
	}

	for _, nodeMachine := range cluster.Nodes {
		if !toDelete[nodeMachine.Node.Name] {
			continue
		}
		machine := nodeMachine.Machine
		node := nodeMachine.Node
		platform.Infof("Deleting %s", node.Name)

		if err := cluster.Cordon(node); err != nil {
			return err
		}
		cluster.Terminate(machine)
	}
	return nil
}

func terminateMaster(platform *platform.Platform, etcdClient *EtcdClient, name string) error {
	ctx := context.TODO()

	// we always interact via the etcd leader, as a previous node may have become unavailable
	leaderClient, err := etcdClient.GetEtcdLeader()
	if err != nil {
		return err
	}
	platform.Infof("etcd leader is: %s", leaderClient.Name)

	members, err := leaderClient.Members(ctx)
	if err != nil {
		return err
	}
	var etcdMember *etcd.Member
	var candidateLeader *etcd.Member
	for _, member := range members {
		if member.Name == name {
			// find the etcd member for the node
			etcdMember = member
		}
		if member.Name != leaderClient.Name {
			// choose a potential candidate to move the etcd leader
			candidateLeader = member
		}
	}
	if etcdMember == nil {
		platform.Warnf("%s has already been removed from etcd cluster", name)
	} else {
		if etcdMember.ID == leaderClient.MemberID {
			platform.Infof("Moving etcd leader from %s to %s", name, candidateLeader.Name)
			if err := leaderClient.MoveLeader(ctx, candidateLeader.ID); err != nil {
				return fmt.Errorf("failed to move leader: %v", err)
			}
		}

		platform.Infof("Removing etcd member %s", name)
		if err := leaderClient.RemoveMember(ctx, etcdMember.ID); err != nil {
			return err
		}
	}

	if platform.Consul != "" {
		// proactively remove server from consul so that we can get a new connection to k8s
		if err := platform.GetConsulClient().RemoveMember(name); err != nil {
			return err
		}
		// reset the connection to the existing master (which may be the one we just removed)
		platform.ResetMasterConnection()
	}
	// wait for a new connection to be healthy before continuing
	if err := platform.WaitFor(); err != nil {
		return err
	}
	return nil
}

// Cleanup stops and deletes all VM's for a cluster;
func Cleanup(platform *platform.Platform) error {
	if platform.TerminationProtection {
		return fmt.Errorf("termination Protection Enabled, use -e terminationProtection=false to disable")
	}

	if err := WithVmwareCluster(platform); err != nil {
		return err
	}
	platform.Terminating = true

	vms, err := platform.Cluster.GetMachines()
	if err != nil {
		return fmt.Errorf("cleanup: failed to get VMs %v", err)
	}

	if len(vms) > platform.GetVMCount()*2 {
		platform.Fatalf("Too many VM's found, expecting +- %d but found %d", platform.GetVMCount(), len(vms))
	}

	platform.Infof("Deleting %d vm's, CTRL+C to skip, sleeping for 10s", len(vms))
	//pausing to give time for user to terminate
	time.Sleep(10 * time.Second)

	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, _vm := range vms {
		vm := _vm
		wg.Add(1)
		go func() {
			defer wg.Done()
			cluster.Terminate(vm)
		}()
	}

	wg.Wait()
	return nil
}
