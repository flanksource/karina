package platform

import (
	"context"
	"fmt"
	"net"
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

func (vm *VM) UUID() string {
	return vm.vm.UUID(context.Background())
}

func (vm *VM) GetVmID() string {
	return vm.vm.Reference().Value
}

// WaitForPoweredOff waits until the VM is reported as off by vCenter
func (vm *VM) WaitForPoweredOff() error {
	return vm.vm.WaitForPowerState(vm.ctx, vim.VirtualMachinePowerStatePoweredOff)
}

func (vm *VM) GetNics(ctx context.Context) ([]vim.GuestNicInfo, error) {
	var nics []vim.GuestNicInfo
	p := property.DefaultCollector(vm.vm.Client())
	err := property.Wait(ctx, p, vm.vm.Reference(), []string{"guest.net"}, func(pc []vim.PropertyChange) bool {
		for _, c := range pc {
			if c.Op != vim.PropertyChangeOpAssign {
				continue
			}

			for _, nic := range c.Val.(vim.ArrayOfGuestNicInfo).GuestNicInfo {
				fmt.Printf("%v %s\n", nic.Connected, nic.Network)
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
func (vm *VM) GetIP(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("Timeout exceeded")
		}
		mo, err := vm.GetVirtualMachine(context.TODO())
		if err != nil {
			return "", err
		}
		if mo.Guest.IpAddress != "" && net.ParseIP(mo.Guest.IpAddress).To4() != nil {
			log.Debugf("[%s] Found IP: %s", vm.Name, mo.Guest.IpAddress)
			return mo.Guest.IpAddress, nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (vm *VM) GetLogicalPortIds(timeout time.Duration) ([]string, error) {
	// deadline := time.Now().Add(timeout)
	ids := []string{}
	devices, err := vm.vm.Device(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, dev := range devices.SelectByType((*vim.VirtualEthernetCard)(nil)) {
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

func (vm *VM) SetAttributes(attributes map[string]string) error {
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

func (vm *VM) GetVirtualMachine(ctx context.Context) (*mo.VirtualMachine, error) {
	var res []mo.VirtualMachine
	pc := property.DefaultCollector(vm.vm.Client())
	err := pc.Retrieve(ctx, []vim.ManagedObjectReference{vm.vm.Reference()}, nil, &res)
	if err != nil {
		return nil, err
	}
	return &(res[0]), nil
}

// WaitForIP waits for a non-local IPv4 address to be reported by vCenter
func (vm *VM) WaitForIP() (string, error) {
	return vm.GetIP(5 * time.Minute)
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
	log.Infof("Gracefully shutting down %s\n", vm.Name)

	err := vm.vm.ShutdownGuest(vm.ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to shutdown %s: %v", vm.Name, err)
	}
	return nil
}

func removeDNS(vm *VM) {
	ip, err := vm.GetIP(time.Second * 5)
	if err != nil {
		log.Warnf("Failed to get IP for %s, unable to remove DNS: %v", vm.Name, err)
		return
	}
	if ip != "" {
		if err := vm.Platform.GetDNSClient().Delete(fmt.Sprintf("*.%s", vm.Platform.Domain), ip); err != nil {
			log.Warnf("Failed to de-register wildcard DNS %s for %s", vm.Name, err)
		}
		if err := vm.Platform.GetDNSClient().Delete(fmt.Sprintf("k8s-api.%s", vm.Platform.Domain), ip); err != nil {
			log.Warnf("Failed to de-register wildcard DNS %s for %s", vm.Name, err)
		}
	}
}

func (vm *VM) Terminate() error {
	log.Infof("Terminating %s", vm.Name)
	if vm.Platform.DryRun {
		log.Infof("Not terminating in dry-run mode %s", vm.Name)
		return nil
	}

	power, _ := vm.vm.PowerState(vm.ctx)
	if power == vim.VirtualMachinePowerStatePoweredOn {
		err := vm.Shutdown()
		if err != nil {
			log.Infof("Graceful shutdown of %s failed, powering off %s\n", vm, err)
			if err := vm.PowerOff(); err != nil {
				log.Infof("Failed to power off %s %s", vm, err)
			}
		}
	} else {
		if err := vm.PowerOff(); err != nil {
			log.Warnf("Failed to power off %s: %v", vm.Name, err)
		}
	}
	vm.vm.WaitForPowerState(vm.ctx, vim.VirtualMachinePowerStatePoweredOff)
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
