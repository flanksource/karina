package provision

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/moshloop/platform-cli/pkg/api"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/nsx"
	"github.com/moshloop/platform-cli/pkg/phases"
	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/types"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	kindCADir = "/etc/flanksource/ingress-ca"
)

// VM provisions a new standalone VM
func VM(platform *platform.Platform, vm *types.VM, konfigs ...string) error {

	if err := platform.OpenViaEnv(); err != nil {
		return fmt.Errorf("vm: failed to open via env: %v", err)
	}
	konfig, err := konfigadm.NewConfig(konfigs...).Build()
	if err != nil {
		return fmt.Errorf("vm: failed to get new config: %v", err)
	}
	log.Infof("Using konfigadm spec: %s\n", konfigs)
	_vm, err := platform.Clone(*vm, konfig)

	if err != nil {
		return fmt.Errorf("vm: failed to clone %v", err)
	}
	log.Infof("Provisioned  %s ->  %s\n", vm.Name, _vm.IP)
	return nil
}

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
			log.Fatalf("Erroring saving config %s", err)
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
	platform.GetKubeConfig()
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}

	masters = platform.GetMasterIPs()
	log.Infof("Detected %d existing masters: %s", len(masters), masters)
	wg := sync.WaitGroup{}
	if platform.Master.Count != len(masters) {
		// upload control plane certs first as secondary masters are created concurrently
		kubeadm.UploadControlPaneCerts(platform)
	}
	for i := 0; i < platform.Master.Count-len(masters); i++ {
		time.Sleep(1 * time.Second)
		wg.Add(1)
		go func() {
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
			wg.Done()
		}()
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

// KindCluster provision or create a kubernetes cluster
func KindCluster(platform *platform.Platform) error {
	kubeadmPatches, err := createKubeAdmPatches(platform)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubeadm patches")
	}

	caPath, err := filepath.Abs(platform.IngressCA.Cert)
	if err != nil {
		return errors.Wrap(err, "failed to expand ca file path")
	}

	kindConfig := kindapi.Cluster{
		TypeMeta: kindapi.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Networking: kindapi.Networking{
			DisableDefaultCNI: true,
		},
		Nodes: []kindapi.Node{
			{
				Role:  "control-plane",
				Image: fmt.Sprintf("kindest/node:%s", platform.Kubernetes.Version),
				ExtraPortMappings: []kindapi.PortMapping{
					{
						ContainerPort: 80,
						HostPort:      80,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
					{
						ContainerPort: 443,
						HostPort:      443,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
					{
						ContainerPort: 6443,
						HostPort:      6443,
						Protocol:      kindapi.PortMappingProtocolTCP,
					},
				},
				KubeadmConfigPatches: kubeadmPatches,
				ExtraMounts: []kindapi.Mount{
					{
						ContainerPath: kindCADir,
						HostPath:      path.Dir(caPath),
						Readonly:      true,
					},
				},
			},
		},
	}

	yml, err := yaml.Marshal(kindConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	tmpfile, err := ioutil.TempFile("", "kind.yml")
	if err != nil {
		return errors.Wrap(err, "failed to create tempfile")
	}
	defer os.Remove(tmpfile.Name())

	if err := ioutil.WriteFile(tmpfile.Name(), yml, 0644); err != nil {
		return errors.Wrap(err, "failed to write kind config file")
	}

	if platform.DryRun {
		fmt.Println(string(yml))
	}

	kind := platform.GetBinary("kind")

	kubeConfig, err := platform.GetKubeConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kube config")
	}

	return kind("create cluster --config %s --kubeconfig %s", tmpfile.Name(), kubeConfig)
}

func provisionLoadbalancers(p *platform.Platform) (masters string, workers string, err error) {

	if p.NSX == nil || p.NSX.Disabled {
		return "", "", nil
	}

	nsxClient, err := p.GetNSXClient()
	if err != nil {
		return "", "", err
	}

	masters, err = nsxClient.CreateLoadBalancer(nsx.LoadBalancerOptions{
		Name:     p.Name + "-masters",
		IPPool:   p.NSX.LoadBalancerIPPool,
		Protocol: "TCP",
		Ports:    []string{"6443"},
		Tier0:    p.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": p.Name + "-masters",
		},
	})
	if err != nil {
		return "", "", err
	}
	workers, err = nsxClient.CreateLoadBalancer(nsx.LoadBalancerOptions{
		Name:     p.Name + "-workers",
		IPPool:   p.NSX.LoadBalancerIPPool,
		Protocol: "TCP",
		Ports:    []string{"80", "443"},
		Tier0:    p.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": p.Name + "-workers",
		},
	})
	if err != nil {
		return "", "", err
	}
	return masters, workers, nil

}

func createKubeAdmPatches(platform *platform.Platform) ([]string, error) {
	clusterConfig := kubeadm.NewClusterConfig(platform)
	clusterConfig.ControlPlaneEndpoint = ""
	clusterConfig.ClusterName = ""
	clusterConfig.APIServer.CertSANs = nil
	clusterConfig.APIServer.ExtraVolumes = []api.HostPathMount{
		{
			Name:      "oidc-certificates",
			HostPath:  path.Join(kindCADir, filepath.Base(platform.IngressCA.Cert)),
			MountPath: "/etc/ssl/oidc/ingress-ca.pem",
			ReadOnly:  true,
			PathType:  api.HostPathFile,
		},
	}
	clusterConfig.APIServer.ExtraArgs["oidc-ca-file"] = "/etc/ssl/oidc/ingress-ca.pem"
	clusterConfig.ControllerManager.ExtraArgs = nil
	clusterConfig.CertificatesDir = ""
	clusterConfig.Networking.PodSubnet = ""
	clusterConfig.Networking.ServiceSubnet = ""

	kubeadmPatches := []interface{}{
		clusterConfig,
		kubeadm.NewInitConfig(platform),
	}

	result := make([]string, len(kubeadmPatches))
	for i, x := range kubeadmPatches {
		yml, err := yaml.Marshal(x)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode yaml for kubeadm patch %v", x)
		}
		result[i] = string(yml)
	}

	return result, nil
}
