package provision

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
	"time"
	// "context"
	"fmt"

	// konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/phases"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func Provision(platform *platform.Platform) error {

	session, err := vmware.GetSessionFromEnv()
	if err != nil {
		return err
	}

	masters := platform.GetMasterIPs()
	vmware.LoadGovcEnvVars(&platform.Master)
	if len(masters) == 0 {
		vm := platform.Master
		vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, "m", utils.ShortTimestamp())

		log.Infof("No  masters detected, deploying new master %s", vm.Name)
		config, err := phases.CreatePrimaryMaster(platform)
		if err != nil {
			log.Fatalf("Failed to create primary master: %s", err)
		}

		data, err := yaml.Marshal(platform)
		if err != nil {
			log.Fatalf("Erroring saving config %s", err)
		}
		err = ioutil.WriteFile("config_saved.yaml", data, 0644)
		if err != nil {
			log.Fatalf("Error saving config: %s", err)
		}

		ip, err := session.Clone(vm, config)

		if err != nil {
			return err
		}
		log.Infof("Provisioned new master: %s\n", ip)

		if err := platform.WaitFor(); err != nil {
			log.Fatalf("Primary master failed to come up %s ", err)
		}
	}

	masters = platform.GetMasterIPs()
	log.Infof("Detected %d existing masters: %s", len(masters), masters)
	wg := sync.WaitGroup{}
	for i := 0; i < platform.Master.Count-len(masters); i++ {
		time.Sleep(1 * time.Second)
		wg.Add(1)
		go func() {
			vm := platform.Master
			vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, "m", utils.ShortTimestamp())
			log.Infof("Creating new secondary master %s\n", vm.Name)
			config, err := phases.CreateSecondaryMaster(platform)
			if err != nil {
				log.Errorf("Failed to create secondary master: %s", err)
			} else {
				ip, err := session.Clone(vm, config)
				if err != nil {
					log.Errorf("Failed to Clone secondary master: %s", err)
				} else {
					log.Infof("Provisioned new master: %s\n", ip)
				}
			}
			wg.Done()
		}()
	}

	for _, worker := range platform.Nodes {
		vmware.LoadGovcEnvVars(&worker)
		for i := 0; i < worker.Count; i++ {
			time.Sleep(1 * time.Second)
			wg.Add(1)
			vm := worker
			go func() {
				config, err := phases.CreateWorker(platform)
				if err != nil {
					log.Errorf("Failed to create workers %s\n", err)
				} else {
					vm.Name = fmt.Sprintf("%s-%s-%s-%s", platform.HostPrefix, platform.Name, "w", utils.ShortTimestamp())
					log.Infof("Creating new worker %s\n", vm.Name)
					ip, err := session.Clone(vm, config)
					if err != nil {
						log.Errorf("Failed to Clone worker: %s", err)
					} else {
						log.Infof("Provisioned new worker: %s\n", ip)
					}
				}
				wg.Done()
			}()

		}
	}
	wg.Wait()
	return nil
}
