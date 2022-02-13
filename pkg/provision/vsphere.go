package provision

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/controller/burnin"
	"github.com/flanksource/karina/pkg/phases"
	"github.com/flanksource/karina/pkg/phases/antrea"
	"github.com/flanksource/karina/pkg/phases/calico"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/provision/vmware"
	"github.com/flanksource/karina/pkg/types"
	"github.com/kr/pretty"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func WithVmwareCluster(p *platform.Platform) error {
	cluster, err := vmware.NewVMwareCluster(p.PlatformConfig)
	if err != nil {
		return err
	}
	p.Cluster = cluster
	if err := p.Init(); err != nil {
		return err
	}

	joinEndpoint, err := p.MasterDiscovery.GetControlPlaneEndpoint(p)
	if err != nil {
		return err
	}
	p.JoinEndpoint = joinEndpoint

	if p.Master.Prefix == "" {
		p.Master.Prefix = "master"
	}

	for name, vm := range p.Nodes {
		if vm.Prefix == "" {
			vm.Prefix = name
		}
		if vm.Cluster == "" {
			vm.Cluster = p.Master.Cluster
		}
		if vm.Template == "" {
			vm.Template = p.Master.Template
		}
		if vm.Folder == "" {
			vm.Folder = p.Master.Folder
		}
		if vm.ContentLibrary == "" {
			vm.ContentLibrary = p.Master.ContentLibrary
		}
		if len(vm.Annotations) == 0 && len(p.Master.Annotations) > 0 {
			vm.Annotations = p.Master.Annotations
		}
		if len(vm.Network) == 0 && len(p.Master.Network) > 0 {
			vm.Network = p.Master.Network
		}
		p.Nodes[name] = vm
	}

	return nil
}

// VsphereCluster provisions or creates a kubernetes cluster
func VsphereCluster(platform *platform.Platform, burninPeriod time.Duration) error {
	if err := WithVmwareCluster(platform); err != nil {
		return err
	}

	if platform.PlatformConfig.Trace {
		pretty.Println(platform.Master) // nolint: errcheck
		pretty.Println(platform.Nodes)  // nolint: errcheck
	}

	if platform.CA == nil {
		return fmt.Errorf("must specify a ca")
	}

	if platform.IngressCA == nil {
		return fmt.Errorf("must specify an ingressCA")
	}

	if platform.Consul == "" && (platform.NSX == nil || platform.NSX.Disabled) && (platform.DNS.Disabled) {
		return fmt.Errorf("must specify a master discovery service e.g. consul, NSX or DNS")
	}

	// first we start a burnin controller in the background that checks
	// new nodes with the burnin taint for health, removing the taint
	// once they become healthy
	burninCancel := make(chan bool)
	go burnin.Run(platform, burninPeriod, burninCancel)
	defer func() {
		burninCancel <- false
	}()

	api, err := platform.GetAPIEndpoint()
	if api == "" || err != nil || !platform.PingMaster() {
		platform.Tracef("No healthy master nodes, creating new master: %v", err)
		// no healthy master endpoint is detected, so we need to create the first control plane node
		// FIXME: Detect situations where all control pane nodes have failed
		_, err := createMaster(platform)
		if err != nil {
			return err
		}
		// after creating the first master we need to deploy the CNI and Cloud Providers to ensure subsequent
		// nodes are tested correctly with the burnin controller
		if err := vsphere.Install(platform); err != nil {
			return err
		}

		if err := calico.Install(platform); err != nil {
			return err
		}

		if err := antrea.Install(platform); err != nil {
			return err
		}

		if err != nil {
			platform.Fatalf("Failed to create master: %v", err)
		}
	}

	masters, err := platform.GetMasterNodes()
	if err != nil {
		return err
	}
	platform.Infof("Detected %d existing masters: %s", len(masters), masters)

	// master nodes are created sequentially due to race conditions when joining etcd
	for i := 0; i < platform.Master.Count-len(masters); i++ {
		_, err := createSecondaryMaster(platform, burninPeriod)
		if err != nil {
			platform.Warnf("Failed to create secondary master: %v", err)
		}
	}

	wg := sync.WaitGroup{}
	existingNodes := platform.GetNodeNames()
	for nodeGroup, worker := range platform.Nodes {
		vms, err := platform.Cluster.GetMachinesFor(&worker)
		if err != nil {
			return err
		}
		missing := []string{}
		for _, vm := range vms {
			if _, ok := existingNodes[vm.Name()]; !ok {
				missing = append(missing, vm.Name())
			}
		}
		for _, m := range missing {
			platform.Errorf("vm did not join kubernetes cluster: %s", m)
			delete(vms, m)
		}

		for i := 0; i < worker.Count-len(vms); i++ {
			time.Sleep(1 * time.Second)
			wg.Add(1)
			_nodeGroup := nodeGroup
			annotations := worker.Annotations
			go func() {
				defer wg.Done()
				if w, err := createWorker(platform, _nodeGroup); err != nil {
					platform.Errorf("Failed to provision worker %v", err)
				} else {
					if err := waitForNode(platform, w.Name(), burninPeriod); err != nil {
						platform.Errorf("%s did not come up healthy, it may need to be re-provisioned %v", w.Name(), err)
					} else {
						if err := addNodeAnnotations(platform, w.Name(), annotations); err != nil {
							platform.Errorf("failed to add annotations to worker %s: %v", w.Name(), err)
						}
						labels := map[string]string{
							constants.NodePoolLabel: nodeGroup,
						}
						if err := addNodeLabels(platform, w.Name(), labels); err != nil {
							platform.Errorf("failed to add labels to worker: %s: %v", w.Name(), err)
						}
					}
				}
			}()
		}
		wg.Wait()
	}

	if err := downscale(platform); err != nil {
		platform.Warnf("failed to downscale: %v ", err)
	}

	endpoint, _ := platform.GetAPIEndpoint()
	fmt.Printf("\n\n\n A new cluster called %s has been provisioned, access it via: https://%s \n\n\n", platform.Name, endpoint)
	return nil
}

func downscale(platform *platform.Platform) error {
	wg := sync.WaitGroup{}
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}
	existingNodes := platform.GetNodeNames()
	for _, worker := range platform.Nodes {
		vms, err := platform.Cluster.GetMachinesFor(&worker)
		if err != nil {
			return err
		}
		missing := []string{}
		for _, vm := range vms {
			if _, ok := existingNodes[vm.Name()]; !ok {
				missing = append(missing, vm.Name())
			}
		}
		for _, m := range missing {
			delete(vms, m)
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
					if err := cluster.Terminate(vm); err != nil {
						platform.Errorf("failed to terminate %s: %v", vm, err)
					}
				}()
			}
		}
	}
	wg.Wait()
	return nil
}

// creating masters needs to be done sequentially due to race conditions in kubeadm
var masterLock sync.Mutex

func createSecondaryMaster(platform *platform.Platform, burninPeriod time.Duration) (types.Machine, error) {
	// upload control plane certs first
	if _, err := kubeadm.UploadControlPlaneCerts(platform); err != nil {
		return nil, err
	}
	masterLock.Lock()
	defer masterLock.Unlock()

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

	platform.Infof("Provisioned new master: %s\n", cloned.IP())

	if err := waitForNode(platform, cloned.Name(), burninPeriod); err != nil {
		platform.Errorf("%s did not come up healthy, it may need to be re-provisioned %v", cloned.Name(), err)
	}

	if err := addNodeAnnotations(platform, cloned.Name(), vm.Annotations); err != nil {
		return nil, errors.Wrapf(err, "failed to add master %s node annotations", cloned.Name())
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
	platform.Infof("No masters detected, deploying new master with %s for master discovery and %s for load balancing", platform.MasterDiscovery, platform.ProvisionHook)
	config, err := phases.CreatePrimaryMaster(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary master: %s", err)
	}

	machine, err := platform.Clone(vm, config)

	if err != nil {
		return nil, err
	}
	platform.Infof("Provisioned new master: %s, waiting for it to become ready", machine.IP())

	// reset any cached connection details
	platform.ResetMasterConnection()
	if err := platform.WaitForAPIServer(); err != nil {
		return nil, fmt.Errorf("primary master failed to come up %s ", err)
	}

	if err := addNodeAnnotations(platform, machine.Name(), vm.Annotations); err != nil {
		return nil, errors.Wrapf(err, "failed to add master %s node annotations", machine.Name())
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

	cloned, err := platform.Clone(vm, config)
	if err != nil {
		return nil, fmt.Errorf("failed to clone worker: %s", err)
	}

	platform.Infof("Provisioned new worker: %s", cloned.IP())
	return cloned, nil
}

// waitForNode waits for the node to become ready and have its burnin taint removed
func waitForNode(platform *platform.Platform, name string, timeout time.Duration) error {
	platform.Infof("[%s] waiting to become ready", name)
	if status, err := platform.WaitForNode(name, timeout, v1.NodeReady, v1.ConditionTrue); err != nil {
		return fmt.Errorf("[%s] did not come up healthy: %v", name, status)
	}

	platform.Infof("[%s] Node has become healthy, waiting for burnin-taint removal", name)
	if err := platform.WaitForTaintRemoval(name, timeout, burnin.Taint); err != nil {
		return fmt.Errorf("[%s] replacement burn-in taint was not removed: %v", name, err)
	}
	return nil
}

func addNodeAnnotations(platform *platform.Platform, name string, annotations map[string]string) error {
	client, err := platform.GetClientset()
	if err != nil {
		return errors.Wrap(err, "failed to get clientset")
	}
	node, err := client.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get node %s", name)
	}

	for k, v := range annotations {
		node.Annotations[k] = v
	}

	if _, err := client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to update node %s", name)
	}
	return nil
}

func addNodeLabels(platform *platform.Platform, name string, labels map[string]string) error {
	client, err := platform.GetClientset()
	if err != nil {
		return errors.Wrap(err, "failed to get clientset")
	}
	node, err := client.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get node %s", name)
	}

	for k, v := range labels {
		node.Labels[k] = v
	}

	if _, err := client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to update node %s", name)
	}
	return nil
}

func backoff(fn func() error, log logger.Logger, backoffOpts *wait.Backoff) error {
	var returnErr *error
	if backoffOpts == nil {
		backoffOpts = &wait.Backoff{
			Duration: 500 * time.Millisecond,
			Factor:   2.0,
			Steps:    7,
		}
	}

	_ = wait.ExponentialBackoff(*backoffOpts, func() (bool, error) {
		err := fn()
		if err == nil {
			return true, nil
		}
		log.Warnf("retrying after error: %v", err)
		returnErr = &err
		return false, nil
	})
	if returnErr != nil {
		return *returnErr
	}
	return nil
}
