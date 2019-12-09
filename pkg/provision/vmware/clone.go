package vmware

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	cloudinit "github.com/moshloop/konfigadm/pkg/cloud-init"
	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	. "github.com/moshloop/platform-cli/pkg/types"
)

const (
	diskMoveType = string(types.VirtualMachineRelocateDiskMoveOptionsMoveAllDiskBackingsAndConsolidate)
)

// Clone kicks off a clone operation on vCenter to create a new virtual machine.
func (s Session) Clone(vm VM, config *konfigadm.Config) (*object.VirtualMachine, error) {
	ctx := context.TODO()

	tpl, err := s.FindVM(vm.Template)
	if err != nil {
		return nil, err
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

	devices, err := tpl.Device(ctx)
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

	cdrom, err := getCdrom(datastore, vm, devices, config)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting cdrom")
	}

	deviceSpecs = append(deviceSpecs, cdrom)

	spec := types.VirtualMachineCloneSpec{
		Config: &types.VirtualMachineConfigSpec{
			Annotation: "Created by platform-cli from " + vm.Template,

			Flags:        newVMFlagInfo(),
			DeviceChange: deviceSpecs,
			NumCPUs:      vm.CPUs,
			MemoryMB:     vm.MemoryGB * 1024,
		},
		Location: types.VirtualMachineRelocateSpec{
			Datastore:    types.NewReference(datastore.Reference()),
			DiskMoveType: diskMoveType,
			Folder:       types.NewReference(folder.Reference()),
			Pool:         types.NewReference(pool.Reference()),
		},
		PowerOn: true,
	}

	log.Infof("Cloning %s to %s", vm.Template, vm.Name)

	task, err := tpl.Clone(ctx, folder, vm.Name, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "error trigging clone op for machine %s/%s %+v\n\n%+v", folder, vm.Name, err, spec)
	}

	_, err = task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	obj, err := s.FindVM(vm.Name)
	if err != nil {
		return nil, err
	}
	log.Infof("Cloned VM: %s", obj.UUID(ctx))
	return obj, nil
}

func newVMFlagInfo() *types.VirtualMachineFlagInfo {
	diskUUIDEnabled := true
	return &types.VirtualMachineFlagInfo{
		DiskUuidEnabled: &diskUUIDEnabled,
	}
}

func getCdrom(datastore *object.Datastore, vm VM, devices object.VirtualDeviceList, config *konfigadm.Config) (types.BaseVirtualDeviceConfigSpec, error) {

	op := types.VirtualDeviceConfigSpecOperationEdit
	cdrom, err := devices.FindCdrom("")
	if err != nil {
		return nil, err
	}

	if cdrom == nil {

		ide, err := devices.FindIDEController("")
		if err != nil {
			return nil, err
		}
		cdrom, err = devices.CreateCdrom(ide)
		if err != nil {
			return nil, err
		}
		op = types.VirtualDeviceConfigSpecOperationAdd
	}
	log.Infof("Creating ISO for %s", vm.Name)
	iso, err := cloudinit.CreateISO(vm.Name, config.ToCloudInit().String())
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("cloud-init/%s.iso", vm.Name)
	log.Infof("Uploading to [%s] %s", datastore.Name(), path)
	if err = datastore.UploadFile(context.TODO(), iso, path, &soap.DefaultUpload); err != nil {
		log.Infof("%+v\n", err)
		return nil, err
	}
	log.Tracef("Uploaded to %s", path)
	cdrom = devices.InsertIso(cdrom, fmt.Sprintf("[%s] %s", vm.Datastore, path))
	devices.Connect(cdrom)
	return &types.VirtualDeviceConfigSpec{
		Operation: op,
		Device:    cdrom,
	}, nil

}
func getDiskSpec(vm VM, devices object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {

	disks := devices.SelectByType((*types.VirtualDisk)(nil))
	if len(disks) != 1 {
		return nil, errors.Errorf("invalid disk count: %d", len(disks))
	}

	disk := disks[0].(*types.VirtualDisk)
	disk.CapacityInKB = int64(vm.DiskGB) * 1024 * 1024

	return &types.VirtualDeviceConfigSpec{
		Operation: types.VirtualDeviceConfigSpecOperationEdit,
		Device:    disk,
	}, nil
}

const ethCardType = "vmxnet3"

func getNetworkSpecs(s Session, vm VM, devices object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	ctx := context.TODO()
	deviceSpecs := []types.BaseVirtualDeviceConfigSpec{}

	// Remove any existing NICs
	for _, dev := range devices.SelectByType((*types.VirtualEthernetCard)(nil)) {
		deviceSpecs = append(deviceSpecs, &types.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: types.VirtualDeviceConfigSpecOperationRemove,
		})
	}

	ref, err := s.Finder.Network(context.TODO(), vm.Network)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find network %q", vm.Network)
	}
	backing, err := ref.EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create new ethernet card backing info for network %q on %q", vm.Network, ctx)
	}
	dev, err := object.EthernetCardTypes().CreateEthernetCard(ethCardType, backing)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create new ethernet card %q for network %q on %q", ethCardType, vm.Network, ctx)
	}

	// Get the actual NIC object. This is safe to assert without a check
	// because "object.EthernetCardTypes().CreateEthernetCard" returns a
	// "types.BaseVirtualEthernetCard" as a "types.BaseVirtualDevice".
	nic := dev.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()

	// Assign a temporary device key to ensure that a unique one will be
	// generated when the device is created.
	nic.Key = -100
	deviceSpecs = append(deviceSpecs, &types.VirtualDeviceConfigSpec{
		Device:    dev,
		Operation: types.VirtualDeviceConfigSpecOperationAdd,
	})
	return deviceSpecs, nil
}
