package provision

import (
	"context"
	"fmt"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	vim "github.com/vmware/govmomi/vim25/types"
	"sync"
)

// Cleanup stops and deletes all VM's for a cluster;
func Cleanup(platform types.PlatformConfig) error {
	ctx := context.TODO()
	session, err := vmware.GetSessionFromEnv()
	if err != nil {
		return err
	}
	list, err := session.Finder.VirtualMachineList(ctx, fmt.Sprintf("%s-%s-*", platform.HostPrefix, platform.Name))
	if err != nil {
		return err
	}

	if len(list) > platform.GetVMCount()*2 {
		log.Fatalf("Too many VM's found, expecting +- %d but found %d", platform.GetVMCount(), len(list))
	}

	var wg sync.WaitGroup
	for _, _vm := range list {
		vm := _vm
		power, _ := vm.PowerState(ctx)
		log.Infof("%s\t%s\t%s\n", vm.Name(), power, platform.Name)
		if platform.DryRun {
			continue
		}
		if power == vim.VirtualMachinePowerStatePoweredOn {
			wg.Add(1)

			go func() {
				log.Infof("Gracefully shutting dow %s\n", vm)
				err := vm.ShutdownGuest(ctx)
				if err != nil {
					log.Infof("Graceful shutdown of %s failed, powering off %s\n", vm, err)
					_, err := vm.PowerOff(ctx)
					if err != nil {
						log.Infof("Failed to power off %s %s", vm, err)
					}
				}
				vm.WaitForPowerState(ctx, vim.VirtualMachinePowerStatePoweredOff)
				log.Infof("Powered off %s\n", vm)
				terminate(ctx, vm, &wg)
			}()
		} else {
			wg.Add(1)
			go func() {
				terminate(ctx, vm, &wg)
			}()
		}
	}
	wg.Wait()
	return nil
}

func terminate(ctx context.Context, vm *object.VirtualMachine, wg *sync.WaitGroup) {
	log.Infof("Terminating  %s\n", vm)
	task, err := vm.Destroy(ctx)
	if err != nil {
		log.Infof("Failed to delete %s %s", vm, err)
		wg.Done()
		return
	}
	info, err := task.WaitForResult(ctx, nil)
	if info.State == "success" {
		log.Infof("Terminated %s\n", vm)
	} else {
		log.Infof("Termination failed %s -> %v\n", vm, info)
	}

	wg.Done()
}
