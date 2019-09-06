package platform

import (
	"context"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	vim "github.com/vmware/govmomi/vim25/types"
)

// VM represents a specific instance of a VM
type VM struct {
	types.VM
	Platform *Platform
	ctx      context.Context
	vm       *object.VirtualMachine
}

func (vm *VM) String() string {
	return vm.Name
}

// WaitForPoweredOff waits until the VM is reported as off by vCenter
func (vm *VM) WaitForPoweredOff() error {
	return vm.vm.WaitForPowerState(vm.ctx, vim.VirtualMachinePowerStatePoweredOff)
}

// WaitForIP waits for a non-local IPv4 address to be reported by vCenter
func (vm *VM) WaitForIP() (string, error) {
	log.Debugf("[%s] Waiting for IP\n", vm)

	ips, err := vm.vm.WaitForNetIP(vm.ctx, true)
	if err != nil {
		return "", nil
	}
	log.Debugf("[%s] Found %s\n", vm, ips)
	for _, ip := range ips {
		return ip[0], nil
	}
	return "", errors.New("Failed to find IP")

}

// Terminate deletes a VM and waits for the destruction to complete (or fail)
func (vm *VM) Terminate() error {
	log.Infof("[%s] terminating\n", vm)
	task, err := vm.vm.Destroy(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete %s", vm)
	}
	info, err := task.WaitForResult(vm.ctx, nil)
	if info.State == "success" {
		log.Debugf("[%s] terminated\n", vm)
	} else {
		return errors.Errorf("Failed to delete %s, %v", vm, info)
	}

	return nil
}

// PowerOff a VM and wait for shutdown to complete,
func (vm *VM) PowerOff() error {
	log.Infof("[%s] powering off\n", vm)
	task, err := vm.vm.PowerOff(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to power off: %s", vm)
	}
	info, err := task.WaitForResult(vm.ctx, nil)
	if info.State == "success" {
		log.Debugf("[%s] powered off\n", vm)
	} else {
		return errors.Errorf("Failed to poweroff %s, %v", vm, info)
	}
	return nil
}

// Shutdown a VM and wait for shutdown to complete,
func (vm *VM) Shutdown() error {
	log.Infof("[%s] shutdown\n", vm)
	err := vm.vm.ShutdownGuest(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to shutdown: %s", vm)
	}
	return nil
}
