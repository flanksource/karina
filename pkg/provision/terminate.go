package provision

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	vim "github.com/vmware/govmomi/vim25/types"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
)

// Cleanup stops and deletes all VM's for a cluster;
func Cleanup(platform *platform.Platform) error {
	ctx := context.TODO()
	session, err := vmware.GetSessionFromEnv()
	if err != nil {
		return err
	}
	prefix := fmt.Sprintf("%s-%s-*", platform.HostPrefix, platform.Name)
	list, err := session.Finder.VirtualMachineList(ctx, prefix)
	if err != nil {
		return err
	}

	if len(list) > platform.GetVMCount()*2 {
		log.Fatalf("Too many VM's found, expecting +- %d but found %d", platform.GetVMCount(), len(list))
	}

	log.Infof("Deleting %d vm's starting with %s, CTRL+C to skip, sleeping for 10s", len(list), prefix)
	//pausing to give time for user to terminate
	time.Sleep(10 * time.Second)

	var wg sync.WaitGroup
	for _, _vm := range list {
		vm := _vm
		log.Infof("Terminating %s", vm.Name())
		if platform.DryRun {
			continue
		}

		wg.Add(1)
		go func() {
			terminate(ctx, platform, vm)
			wg.Done()
		}()

	}
	wg.Wait()
	return nil
}

func terminate(ctx context.Context, platform *platform.Platform, vm *object.VirtualMachine) {
	ips, err := vm.WaitForNetIP(context.TODO(), true)
	ip := []string{}
	for _, _ip := range ips {
		ip = append(ip, _ip...)
	}
	if err != nil {
		log.Warnf("Failed to get IP for %s: %v", vm.Name(), err)
	}
	if platform.DryRun {
		log.Infof("Not terminating in dry-run mode %s", vm.Name())
		return
	}
	if len(ips) > 0 {
		if err := platform.GetDNSClient().Delete(fmt.Sprintf("*.%s", platform.Domain), ip...); err != nil {
			log.Warnf("Failed to de-register wildcard DNS %s for %s", vm.Name, err)
		}
		if err := platform.GetDNSClient().Delete(fmt.Sprintf("k8s-api.%s", platform.Domain), ip...); err != nil {
			log.Warnf("Failed to de-register wildcard DNS %s for %s", vm.Name, err)
		}
	}
	power, _ := vm.PowerState(ctx)

	if power == vim.VirtualMachinePowerStatePoweredOn {
		log.Infof("Gracefully shutting down %s\n", vm)
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
	}

	log.Infof("Terminating  %s\n", vm)
	task, err := vm.Destroy(ctx)
	if err != nil {
		log.Infof("Failed to delete %s %s", vm, err)
		return
	}
	info, err := task.WaitForResult(ctx, nil)
	if info.State == "success" {
		log.Infof("Terminated %s\n", vm)
	} else {
		log.Infof("Termination failed %s -> %v\n", vm, info)
	}
}
