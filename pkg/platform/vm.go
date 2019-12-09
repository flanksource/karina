package platform

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	vim "github.com/vmware/govmomi/vim25/types"

	"github.com/moshloop/platform-cli/pkg/types"
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
func (vm *VM) GetIP(timeout time.Duration) (string, error) {
	deadline, cancel := context.WithDeadline(context.TODO(), time.Now().Add(timeout))
	defer cancel()
	ips, err := vm.vm.WaitForNetIP(deadline, true)
	if err != nil {
		return "", nil
	}
	log.Tracef("[%s] Found %s\n", vm, ips)
	for _, ip := range ips {
		return ip[0], nil
	}
	return "", errors.New("Failed to find IP")
}

func (vm *VM) SetAttribtues(attributes map[string]string) error {
	ctx := context.TODO()
	fields, err := object.GetCustomFieldsManager(vm.vm.Client())
	if err != nil {
		return err
	}
	for k, v := range attributes {
		key, err := fields.FindKey(ctx, k)
		if err != nil {
			return err
		}
		if err := fields.Set(ctx, vm.vm.Reference(), key, v); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VM) GetAttributes() (map[string]string, error) {
	attributes := make(map[string]string)
	ctx := context.TODO()
	refs := []vim.ManagedObjectReference{vm.vm.Reference()}
	var objs []mo.ManagedEntity
	if err := property.DefaultCollector(vm.vm.Client()).Retrieve(ctx, refs, []string{"name", "customValue"}, &objs); err != nil {
		return nil, err
	}

	fields, err := object.GetCustomFieldsManager(vm.vm.Client())
	if err != nil {
		return nil, err
	}
	field, err := fields.Field(ctx)
	if err != nil {
		return nil, err
	}

	for _, obj := range objs {
		for i := range obj.CustomValue {
			val := obj.CustomValue[i].(*vim.CustomFieldStringValue)
			attributes[field.ByKey(val.Key).Name] = val.Value
		}
	}

	return attributes, nil
}

// WaitForIP waits for a non-local IPv4 address to be reported by vCenter
func (vm *VM) WaitForIP() (string, error) {
	return vm.GetIP(5 * time.Minute)
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
