package provision

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		if err := cluster.Terminate(orphan); err != nil {
			platform.Errorf("failed to terminate %s: %v", orphan, err)
		}
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
		if err := cluster.Terminate(machine); err != nil {
			platform.Errorf("failed to terminate %s: %v", machine, err)
		}
	}
	return nil
}

func terminate(platform *platform.Platform, vm types.Machine) {
	if err := platform.ProvisionHook.BeforeTerminate(platform, vm); err != nil {
		platform.Warnf("[%s] failed to call before terminate: %v", vm, err)
		return
	}

	if !platform.Terminating {
		if err := platform.Drain(vm.Name(), 2*time.Minute); err != nil {
			platform.Warnf("[%s] failed to drain: %v", vm.Name(), err)
		}
	}

	client, err := platform.GetClientset()
	if err != nil {
		platform.Warnf("[%s] failed to get client to delete node: %v", vm, err)
	} else {
		node, err := client.CoreV1().Nodes().Get(context.TODO(), vm.Name(), metav1.GetOptions{})
		if err != nil {
			// we always attempt to terminate a node as a master if we don't know
			// to ensure it is always removed from etcd
			if err := terminateMaster(platform, vm.Name()); err != nil {
				platform.Warnf("Failed to terminate master %v", err)
			}
		} else if kommons.IsMasterNode(*node) {
			if err := terminateMaster(platform, node.Name); err != nil {
				platform.Warnf("Failed to terminate master %v", err)
			}
		}
		if err := backoff(func() error {
			return platform.DeleteNode(vm.Name())
		}, platform.Logger, nil); err != nil {
			platform.Warnf("[%s] failed to delete node: %v", vm, err)
		}
	}

	if err := platform.ProvisionHook.AfterTerminate(platform, vm); err != nil {
		platform.Warnf("[%s] failed calling AfterTerminate hook: %v", vm, err)
	}

	if err := vm.Terminate(); err != nil {
		platform.Warnf("[%s] failed to terminate %s: %v", vm.Name(), err)
	}
}

func terminateConsul(platform *platform.Platform, name string) error {
	if platform.Consul != "" {
		// proactively remove server from consul so that we can get a new connection to k8s
		if err := platform.GetConsulClient().RemoveMember(name); err != nil {
			return err
		}
	}
	return nil
}

func terminateMaster(platform *platform.Platform, name string) error {
	if err := backoff(func() error {
		return terminateConsul(platform, name)
	}, platform.Logger, nil); err != nil {
		return err
	}

	// reset the connection to the existing master (which may be the one we just removed)
	platform.ResetMasterConnection()

	// wait for a new connection to be healthy before continuing
	if err := platform.WaitForAPIServer(); err != nil {
		return err
	}
	return nil
}

// Terminate stops and deletes all VM's for a cluster;
func Terminate(platform *platform.Platform) error {
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
			if err := cluster.Terminate(vm); err != nil {
				platform.Errorf("failed to terminate %s: %v", vm, err)
			}
		}()
	}

	wg.Wait()
	return nil
}
