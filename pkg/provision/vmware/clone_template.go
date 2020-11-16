package vmware

import (
	"context"
	"fmt"
	"net/url"
	"os"

	ptypes "github.com/flanksource/karina/pkg/types"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/types"
)

// CloneTemplate creates a new VM from a content library template
func (s Session) CloneTemplate(vm ptypes.VM, config *konfigadm.Config) (*object.VirtualMachine, error) {
	ctx := context.TODO()

	libraryItem, err := s.FindTemplate(vm.ContentLibrary, vm.Template)
	if err != nil {
		return nil, fmt.Errorf("getVirtualMachine: retrieve failed: %v", err)
	}

	folder, err := s.Finder.FolderOrDefault(ctx, vm.Folder)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get folder for %q", ctx)
	}

	datastore, err := s.Finder.DatastoreOrDefault(ctx, vm.Datastore)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get datastore for %q", ctx)
	}

	pool, err := s.Finder.ResourcePoolOrDefault(ctx, vm.ResourcePool)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get resource pool for %q", ctx)
	}

	restClient := rest.NewClient(s.Client.Client)
	user := url.UserPassword(os.Getenv("GOVC_USER"), os.Getenv("GOVC_PASS"))
	if err := restClient.Login(context.TODO(), user); err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}
	m := vcenter.NewManager(restClient)

	if libraryItem.Type != "ovf" {
		return nil, fmt.Errorf("library item type %s not supported", libraryItem.Type)
	}

	deploy := vcenter.Deploy{
		DeploymentSpec: vcenter.DeploymentSpec{
			Name:               vm.Name,
			DefaultDatastoreID: datastore.Reference().Value,
			AcceptAllEULA:      true,
			Annotation:         "Created by karina from " + vm.Template,
		},
		Target: vcenter.Target{
			ResourcePoolID: pool.Reference().Value,
			FolderID:       folder.Reference().Value,
		},
	}

	s.Infof("Deploying library item %s to %s", libraryItem.Name, vm.Name)

	_, err = m.DeployLibraryItem(ctx, libraryItem.ID, deploy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy library item")
	}

	obj, err := s.FindVM(vm.Name)
	if err != nil {
		return nil, fmt.Errorf("clone: failed to find VM: %v", err)
	}
	s.Infof("Cloned VM: %s", obj.UUID(ctx))

	s.Infof("Configuring VM: %s", vm.Name)

	devices, err := obj.Device(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting devices for %q", ctx)
	}
	deviceSpecs := []types.BaseVirtualDeviceConfigSpec{}
	if vm.DiskGB > 0 {
		diskSpec, err := getDiskSpec(vm, devices)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting disk spec for %q", ctx)
		}
		deviceSpecs = append(deviceSpecs, diskSpec)
	}

	networkSpecs, err := getNetworkSpecs(s, vm, devices)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting network specs for %q", ctx)
	}
	deviceSpecs = append(deviceSpecs, networkSpecs...)

	cdrom, err := s.getCdrom(datastore, vm, devices, config)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting cdrom")
	}
	deviceSpecs = append(deviceSpecs, cdrom)

	serial, err := s.getSerial(vm, devices)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting serial device")
	}
	deviceSpecs = append(deviceSpecs, serial)

	spec := types.VirtualMachineConfigSpec{
		Annotation:   "Created by karina from " + vm.Template,
		Flags:        newVMFlagInfo(),
		DeviceChange: deviceSpecs,
		NumCPUs:      vm.CPUs,
		MemoryMB:     vm.MemoryGB * 1024,
	}

	task, err := obj.Reconfigure(ctx, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to reconfigure vm %s", vm.Name)
	}

	_, err = task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for reconfigure for vm %s", vm.Name)
	}

	task, err = obj.PowerOn(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to power on")
	}

	_, err = task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for power on")
	}

	obj, err = s.FindVM(vm.Name)
	if err != nil {
		return nil, fmt.Errorf("clone: failed to find VM: %v", err)
	}

	return obj, nil
}
