package provision

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/moshloop/platform-cli/pkg/phases"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
)

// VsphereCluster provision or create a kubernetes cluster
func VsphereCluster(platform *platform.Platform) error {
	if err := platform.OpenViaEnv(); err != nil {
		log.Fatalf("Failed to initialize platform: %s", err)
	}

	masters := platform.GetMasterIPs()
	vmware.LoadGovcEnvVars(&platform.Master)
	if len(masters) == 0 {
		vm := platform.Master
		vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, "m", utils.ShortTimestamp())
		vm.Tags["Role"] = platform.Name + "-masters"
		log.Infof("No masters detected, deploying new master %s", vm.Name)
		config, err := phases.CreatePrimaryMaster(platform)
		if err != nil {
			log.Fatalf("Failed to create primary master: %s", err)
		}

		data, err := yaml.Marshal(platform.PlatformConfig)
		if err != nil {
			log.Fatalf("Error saving config %s", err)
		}

		log.Tracef("Using configuration: \n%s\n", console.StripSecrets(string(data)))

		if !platform.DryRun {
			vm, err := platform.Clone(vm, config)

			if err != nil {
				return err
			}
			if err := platform.GetDNSClient().Append(fmt.Sprintf("k8s-api.%s", platform.Domain), vm.IP); err != nil {
				return err
			}
			log.Infof("Provisioned new master: %s, waiting for it to become ready", vm.IP)
		}
		if err := platform.WaitFor(); err != nil {
			log.Fatalf("Primary master failed to come up %s ", err)
		}
	}

	// make sure admin kubeconfig is available
	platform.GetKubeConfig() // nolint: errcheck
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}

	masters = platform.GetMasterIPs()
	log.Infof("Detected %d existing masters: %s", len(masters), masters)
	wg := sync.WaitGroup{}
	if platform.Master.Count != len(masters) {
		// upload control plane certs first
		kubeadm.UploadControlPaneCerts(platform) // nolint: errcheck
	}
	for i := 0; i < platform.Master.Count-len(masters); i++ {
		vm := platform.Master
		vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, vm.Prefix, utils.ShortTimestamp())
		vm.Tags["Role"] = platform.Name + "-masters"
		log.Infof("Creating new secondary master %s\n", vm.Name)
		config, err := phases.CreateSecondaryMaster(platform)
		if err != nil {
			log.Errorf("Failed to create secondary master: %s", err)
		} else {
			if !platform.DryRun {
				vm, err := platform.Clone(vm, config)
				if err != nil {
					log.Errorf("Failed to Clone secondary master: %s", err)
				} else {
					if err := platform.GetDNSClient().Append(fmt.Sprintf("k8s-api.%s", platform.Domain), vm.IP); err != nil {
						log.Warnf("Failed to update DNS for %s", vm.IP)
					} else {
						log.Infof("Provisioned new master: %s\n", vm.IP)
					}
				}
			}
		}
	}

	for _, worker := range platform.Nodes {
		vmware.LoadGovcEnvVars(&worker)

		vms, err := platform.GetVMsByPrefix(worker.Prefix)
		if err != nil {
			return err
		}

		for i := 0; i < worker.Count-len(vms); i++ {
			time.Sleep(1 * time.Second)
			wg.Add(1)
			vm := worker
			go func() {
				config, err := phases.CreateWorker(platform)
				if err != nil {
					log.Errorf("Failed to create workers %s\n", err)
				} else {
					vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, worker.Prefix, utils.ShortTimestamp())
					vm.Tags["Role"] = platform.Name + "-workers"
					if !platform.DryRun {
						log.Infof("Creating new worker %s\n", vm.Name)
						vm, err := platform.Clone(vm, config)
						if err != nil {
							log.Errorf("Failed to Clone worker: %s", err)
						} else {
							if err := platform.GetDNSClient().Append(fmt.Sprintf("*.%s", platform.Domain), vm.IP); err != nil {
								log.Warnf("Failed to update DNS for %s", vm.IP)
							} else {
								log.Infof("Provisioned new worker: %s\n", vm.IP)
							}
						}
					}
				}
				wg.Done()
			}()
		}

		if worker.Count < len(vms) {
			terminateCount := len(vms) - worker.Count
			var vmNames []string
			for k := range vms {
				vmNames = append(vmNames, k)
			}
			sort.Strings(vmNames)
			log.Infof("Downscaling %d extra worker nodes", terminateCount)
			time.Sleep(3 * time.Second)
			for i := 0; i < terminateCount; i++ {
				name := vmNames[worker.Count+i-1]
				wg.Add(1)
				go func() {
					//terminate oldest first
					if err := vms[name].Terminate(); err != nil {
						log.Warnf("Failed to terminate %s: %v", name, err)
					}
					wg.Done()
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
		log.Errorf("Failed to provision load balancers: %v", err)
	}
	fmt.Printf("Provisioned LoadBalancers:\n Masters: %s\nWorkers: %s\n", masterLB, workerLB)
	return nil
}
