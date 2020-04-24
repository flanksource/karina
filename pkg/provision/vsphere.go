package provision

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/yaml"
	"github.com/moshloop/platform-cli/pkg/phases"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WithVmwareCluster(platform *platform.Platform) error {
	cluster, err := vmware.NewVMwareCluster(platform.HostPrefix + "-" + platform.Name)
	if err != nil {
		return err
	}
	platform.Cluster = cluster
	return nil
}

// VsphereCluster provisions or creates a kubernetes cluster
func VsphereCluster(platform *platform.Platform) error {
	if err := WithVmwareCluster(platform); err != nil {
		return err
	}

	masters := platform.GetMasterIPs()
	if len(masters) == 0 {
		_, err := createMaster(platform)
		if err != nil {
			platform.Fatalf("Failed to create master: %v", err)
		}
	}

	// make sure admin kubeconfig is available
	platform.GetKubeConfig() // nolint: errcheck
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}

	masters = platform.GetMasterIPs()
	platform.Infof("Detected %d existing masters: %s", len(masters), masters)

	if platform.Master.Count != len(masters) {
		// upload control plane certs first
		kubeadm.UploadControlPaneCerts(platform) // nolint: errcheck
	}
	for i := 0; i < platform.Master.Count-len(masters); i++ {
		_, err := createSecondaryMaster(platform)
		if err != nil {
			platform.Warnf("Failed to create secondary master: %v", err)
		}
	}

	wg := sync.WaitGroup{}
	for nodeGroup, worker := range platform.Nodes {
		vms, err := platform.Cluster.GetMachinesByPrefix(worker.Prefix)
		if err != nil {
			return err
		}

		for i := 0; i < worker.Count-len(vms); i++ {
			time.Sleep(1 * time.Second)
			wg.Add(1)
			_nodeGroup := nodeGroup
			go func() {
				defer wg.Done()
				if _, err := createWorker(platform, _nodeGroup); err != nil {
					platform.Errorf("Failed to provision worker %v", err)
				}
			}()
		}

		if worker.Count < len(vms) {
			terminateCount := len(vms) - worker.Count
			var vmNames []string
			for k := range vms {
				vmNames = append(vmNames, k)
			}
			sort.Strings(vmNames)
			platform.Infof("Downscaling %d extra worker nodes", terminateCount)
			time.Sleep(3 * time.Second)
			for i := 0; i < terminateCount; i++ {
				vm := vms[vmNames[worker.Count+i-1]] //terminate oldest first
				wg.Add(1)
				go func() {
					defer wg.Done()
					terminate(platform, vm)
				}()
			}
		}
	}
	wg.Wait()

	path, err := platform.GetKubeConfig()
	if err != nil {
		return err
	}
	fmt.Printf("\n\n\n A new cluster called %s has been provisioned, access it via: kubectl --kubeconfig %s get nodes\n\n Next deploy the CNI and addons\n\n\n", platform.Name, path)
	masterLB, workerLB, err := provisionLoadbalancers(platform)
	if err != nil {
		platform.Errorf("Failed to provision load balancers: %v", err)
	}
	fmt.Printf("Provisioned LoadBalancers:\n Masters: %s\nWorkers: %s\n", masterLB, workerLB)
	return nil
}

func createSecondaryMaster(platform *platform.Platform) (types.Machine, error) {
	vm := platform.Master
	vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, vm.Prefix, utils.ShortTimestamp())
	if vm.Tags == nil {
		vm.Tags = make(map[string]string)
	}
	vm.Tags["Role"] = platform.Name + "-masters"
	platform.Infof("Creating new secondary master %s", vm.Name)
	if platform.DryRun {
		return nil, nil
	}
	config, err := phases.CreateSecondaryMaster(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to create secondary master: %s", err)
	}
	cloned, err := platform.Clone(vm, config)
	if err != nil {
		return nil, fmt.Errorf("failed to clone secondary master: %s", err)
	}
	if err := platform.GetDNSClient().Append(fmt.Sprintf("k8s-api.%s", platform.Domain), cloned.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", cloned.IP())
	} else {
		platform.Infof("Provisioned new master: %s\n", cloned.IP())
	}
	return cloned, nil
}

func createMaster(platform *platform.Platform) (types.Machine, error) {
	vm := platform.Master
	vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, "m", utils.ShortTimestamp())
	if vm.Tags == nil {
		vm.Tags = make(map[string]string)
	}
	vm.Tags["Role"] = platform.Name + "-masters"
	platform.Infof("No masters detected, deploying new master %s", vm.Name)
	config, err := phases.CreatePrimaryMaster(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary master: %s", err)
	}

	data, err := yaml.Marshal(platform.PlatformConfig)
	if err != nil {
		return nil, fmt.Errorf("error saving config %s", err)
	}

	platform.Tracef("Using configuration: \n%s\n", console.StripSecrets(string(data)))

	var machine types.Machine
	if !platform.DryRun {
		//Note: = not :=, otherwise the new `machine` shadows the one declared
		//                outside the if and this function always return nil
		machine, err = platform.Clone(vm, config)

		if err != nil {
			return nil, err
		}
		if err := platform.GetDNSClient().Append(fmt.Sprintf("k8s-api.%s", platform.Domain), machine.IP()); err != nil {
			platform.Errorf("Failed to update DNS record for %s: %v", machine, err)
		}
		platform.Infof("Provisioned new master: %s, waiting for it to become ready", machine.IP())
	}
	if err := platform.WaitFor(); err != nil {
		return nil, fmt.Errorf("primary master failed to come up %s ", err)
	}
	return machine, nil
}

func createWorker(platform *platform.Platform, nodeGroup string) (types.Machine, error) {
	if nodeGroup == "" {
		for k := range platform.Nodes {
			nodeGroup = k
		}
	}
	worker := platform.Nodes[nodeGroup]
	vm := worker
	config, err := phases.CreateWorker(nodeGroup, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker %v", err)
	}
	vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, vm.Prefix, utils.ShortTimestamp())
	if vm.Tags == nil {
		vm.Tags = make(map[string]string)
	}
	vm.Tags["Role"] = platform.Name + "-workers"
	platform.Infof("Creating new worker %s", vm.Name)
	if platform.DryRun {
		return nil, nil
	}

	cloned, err := platform.Clone(vm, config)
	if err != nil {
		return nil, fmt.Errorf("failed to clone worker: %s", err)
	}
	if err := platform.GetDNSClient().Append(fmt.Sprintf("*.%s", platform.Domain), cloned.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", cloned.IP())
	} else {
		platform.Infof("Provisioned new worker: %s\n", cloned.IP())
	}
	return cloned, nil
}

func terminate(platform *platform.Platform, vm types.Machine) {
	if !platform.Terminating {
		if err := platform.Drain(vm.Name(), 2*time.Minute); err != nil {
			platform.Warnf("[%s] failed to drain: %v", vm.Name(), err)
		}
	}
	client, err := platform.GetClientset()
	if err != nil {
		platform.Warnf("Failed to get client to delete node")
	} else {
		if err := client.CoreV1().Nodes().Delete(vm.Name(), &metav1.DeleteOptions{}); err != nil {
			platform.Warnf("Failed to delete node for %s: %v", vm, err)
		}
	}

	if err := RemoveDNS(platform, vm); err != nil {
		platform.Warnf("Failed to remove dns for %s: %v", vm, err)
	}
	if err := vm.Terminate(); err != nil {
		platform.Warnf("Failed to terminate %s: %v", vm.Name(), err)
	}
}

func RemoveDNS(p *platform.Platform, vm types.Machine) error {
	ip, err := vm.GetIP(time.Second * 5)
	if err != nil {
		return fmt.Errorf("failed to get IP for %s, unable to remove DNS: %v", vm, err)
	}
	if ip != "" {
		if err := p.GetDNSClient().Delete(fmt.Sprintf("*.%s", p.Domain), ip); err != nil {
			return err
		}
		if err := p.GetDNSClient().Delete(fmt.Sprintf("k8s-api.%s", p.Domain), ip); err != nil {
			return err
		}
	}
	return nil
}
