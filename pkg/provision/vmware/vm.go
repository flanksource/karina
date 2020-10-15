package vmware

import (
	"context"
	"fmt"

	"net"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	vim "github.com/vmware/govmomi/vim25/types"

	"github.com/flanksource/karina/pkg/types"
	"github.com/vmware/govmomi/vapi/rest"
	vtags "github.com/vmware/govmomi/vapi/tags"
)

// VM represents a specific instance of a VM
type vm struct {
	logger.Logger
	name, ip string
	dryRun   bool
	ctx      context.Context
	config   *types.VM
	vm       *object.VirtualMachine
}

func NewVM(ctx context.Context, dryRun bool, obj *object.VirtualMachine, config *types.VM) types.Machine {
	_vm := vm{
		Logger: logger.WithValues("vm", obj.Name()),
		ctx:    ctx,
		dryRun: dryRun,
		vm:     obj,
		config: config,
		name:   obj.Name(),
	}
	return &_vm
}

func (vm *vm) GetTags() map[string]string {
	return vm.config.Tags
}

func (vm *vm) IP() string {
	if vm.ip == "" {
		ip, _ := vm.WaitForIP()
		vm.ip = ip
	}
	return vm.ip
}

func (vm *vm) Name() string {
	return vm.name
}

func (vm *vm) GetAge() time.Duration {
	attributes, _ := vm.GetAttributes()
	created, _ := time.ParseInLocation("02Jan06-15:04:05", attributes["CreatedDate"], time.Local)
	return time.Since(created)
}

func (vm *vm) GetTemplate() string {
	attributes, _ := vm.GetAttributes()
	return attributes["Template"]
}

func (vm *vm) String() string {
	return vm.name
}

func (vm *vm) UUID() string {
	return vm.vm.UUID(context.Background())
}

// nolint: golint, stylecheck
func (vm *vm) GetVmID() string {
	return vm.vm.Reference().Value
}

// WaitForPoweredOff waits until the VM is reported as off by vCenter
func (vm *vm) WaitForPoweredOff() error {
	return vm.vm.WaitForPowerState(vm.ctx, vim.VirtualMachinePowerStatePoweredOff)
}

func (vm *vm) GetNics(ctx context.Context) ([]vim.GuestNicInfo, error) {
	var nics []vim.GuestNicInfo
	p := property.DefaultCollector(vm.vm.Client())
	err := property.Wait(ctx, p, vm.vm.Reference(), []string{"guest.net"}, func(pc []vim.PropertyChange) bool {
		for _, c := range pc {
			if c.Op != vim.PropertyChangeOpAssign {
				continue
			}

			for _, nic := range c.Val.(vim.ArrayOfGuestNicInfo).GuestNicInfo {
				mac := nic.MacAddress
				if mac == "" || nic.IpConfig == nil {
					continue
				}

				for _, ip := range nic.IpConfig.IpAddress {
					if net.ParseIP(ip.IpAddress).To4() == nil {
						continue // Ignore non IPv4 address
					}
					nics = append(nics, nic)
				}
			}
		}
		return len(nics) == 0
	})
	return nics, err
}

// WaitForIP waits for a non-local IPv4 address to be reported by vCenter
func (vm *vm) GetIP(timeout time.Duration) (string, error) {
	if !vm.IsPoweredOn() {
		return "<powered off>", nil
	}
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout exceeded")
		}
		mo, err := vm.GetVirtualMachine(context.TODO())
		if err != nil {
			return "", err
		}
		if mo.Guest.IpAddress != "" && net.ParseIP(mo.Guest.IpAddress).To4() != nil {
			vm.Debugf("Found IP: %s", mo.Guest.IpAddress)
			return mo.Guest.IpAddress, nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (vm *vm) GetLogicalPortIds(timeout time.Duration) ([]string, error) {
	// deadline := time.Now().Add(timeout)
	ids := []string{}
	devices, err := vm.vm.Device(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("getLogicalPortIds: failed to get devices: %v", err)
	}

	for _, dev := range devices.SelectByType((*vim.VirtualEthernetCard)(nil)) {
		// nolint: gosimple
		switch dev.(type) {
		case *vim.VirtualVmxnet3:
			net := dev.(*vim.VirtualVmxnet3)
			ids = append(ids, net.ExternalId)
		case *vim.VirtualEthernetCard:
			ids = append(ids, dev.(*vim.VirtualEthernetCard).ExternalId)
		}
	}
	return ids, nil
}

func (vm *vm) SetAttributes(attributes map[string]string) error {
	ctx := context.TODO()
	fields, err := object.GetCustomFieldsManager(vm.vm.Client())
	if err != nil {
		return fmt.Errorf("setAttributes: failed to get custom field manager: %v", err)
	}
	for k, v := range attributes {
		key, err := fields.FindKey(ctx, k)
		if err != nil {
			return fmt.Errorf("setAttributes: failed to find key: %v", err)
		}
		if err := fields.Set(ctx, vm.vm.Reference(), key, v); err != nil {
			return fmt.Errorf("setAttributes: failed to set fields: %v", err)
		}
	}
	return nil
}

func (vm *vm) GetAttributes() (map[string]string, error) {
	attributes := make(map[string]string)
	ctx := context.TODO()
	refs := []vim.ManagedObjectReference{vm.vm.Reference()}
	var objs []mo.ManagedEntity
	if err := property.DefaultCollector(vm.vm.Client()).Retrieve(ctx, refs, []string{"name", "customValue"}, &objs); err != nil {
		return nil, fmt.Errorf("getAttributes: failed to set default collector: %v", err)
	}

	fields, err := object.GetCustomFieldsManager(vm.vm.Client())
	if err != nil {
		return nil, fmt.Errorf("getAttributes: failed to get fields: %v", err)
	}
	field, err := fields.Field(ctx)
	if err != nil {
		return nil, fmt.Errorf("getAttributes: failed to get field: %v", err)
	}

	for _, obj := range objs {
		for i := range obj.CustomValue {
			val := obj.CustomValue[i].(*vim.CustomFieldStringValue)
			attributes[field.ByKey(val.Key).Name] = val.Value
		}
	}

	return attributes, nil
}

func (vm *vm) GetVirtualMachine(ctx context.Context) (*mo.VirtualMachine, error) {
	var res []mo.VirtualMachine
	pc := property.DefaultCollector(vm.vm.Client())
	err := pc.Retrieve(ctx, []vim.ManagedObjectReference{vm.vm.Reference()}, nil, &res)
	if err != nil {
		return nil, fmt.Errorf("getVirtualMachine: retrieve failed: %v", err)
	}
	return &(res[0]), nil
}

// WaitForIP waits for a non-local IPv4 address to be reported by vCenter
func (vm *vm) WaitForIP() (string, error) {
	return vm.GetIP(5 * time.Minute)
}

// PowerOff a VM and wait for shutdown to complete,
func (vm *vm) PowerOff() error {
	vm.Infof("powering off")
	task, err := vm.vm.PowerOff(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to power off: %s", vm)
	}
	info, err := task.WaitForResult(vm.ctx, nil)
	if info.State == "success" {
		vm.Debugf("powered off")
	} else {
		return errors.Errorf("Failed to poweroff %s, %v, %s", vm, info, err)
	}
	return nil
}

// Shutdown a VM and wait for shutdown to complete,
func (vm *vm) Shutdown() error {
	vm.Infof("gracefully shutting down")
	err := vm.vm.ShutdownGuest(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to shutdown %s: %v", vm.Name(), err)
	}
	return nil
}

func (vm *vm) IsPoweredOn() bool {
	power, _ := vm.vm.PowerState(vm.ctx)
	return power == vim.VirtualMachinePowerStatePoweredOn
}

func (vm *vm) Terminate() error {
	vm.Infof("terminating")
	if vm.dryRun {
		vm.Infof("Not terminating in dry-run mode")
		return nil
	}

	power, _ := vm.vm.PowerState(vm.ctx)
	if power == vim.VirtualMachinePowerStatePoweredOn {
		err := vm.Shutdown()
		if err != nil {
			vm.Infof("graceful shutdown failed, powering off %s", err)
			if err := vm.PowerOff(); err != nil {
				vm.Infof("failed to power off: %s", err)
			}
		}
	} else {
		if err := vm.PowerOff(); err != nil {
			vm.Warnf("failed to power off %v", err)
		}
	}
	_ = vm.vm.WaitForPowerState(vm.ctx, vim.VirtualMachinePowerStatePoweredOff)
	task, err := vm.vm.Destroy(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete %s", vm)
	}
	info, err := task.WaitForResult(vm.ctx, nil)
	if info.State == "success" {
		vm.Debugf("terminated")
	} else {
		return errors.Errorf("Failed to delete %s, %v, %s", vm, info, err)
	}
	return nil
}

func (vm *vm) SetTags(tags map[string]string) error {
	message := "Setting tags ["
	for k, v := range tags {
		message += fmt.Sprintf("%s=%s ", k, v)
	}
	message += fmt.Sprintf("] to virtual machine %s", vm.name)
	vm.Infof(message)

	restClient := rest.NewClient(vm.vm.Client())
	manager := vtags.NewManager(restClient)

	for categoryID, tagName := range tags {
		categoryTags, err := manager.GetTagsForCategory(vm.ctx, categoryID)
		if err != nil {
			return errors.Wrapf(err, "failed to list tags for category %s: %v", categoryID, err)
		}
		tagID := ""
		for _, t := range categoryTags {
			if t.Name == tagName {
				tagID = t.ID
			}
		}

		if tagID != "" {
			manager.AttachTag(vm.ctx, tagID, vm.vm.Reference())
		}
	}

	return nil
}
