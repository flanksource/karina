package vmware

import (
	"context"
	"fmt"
	ptypes "github.com/flanksource/karina/pkg/types"
	cloudinit "github.com/flanksource/konfigadm/pkg/cloud-init"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"io/ioutil"
	"os"
)

const (
	diskMoveType = string(types.VirtualMachineRelocateDiskMoveOptionsMoveAllDiskBackingsAndConsolidate)
)

// Clone kicks off a clone operation on vCenter to create a new virtual machine.
func (s Session) Clone(vm ptypes.VM, config *konfigadm.Config) (*object.VirtualMachine, error) {
	ctx := context.TODO()

	tpl, err := s.FindVM(vm.Template)
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

	cdrom, err := s.getCdrom(datastore, vm, devices, config)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting cdrom")
	}
	deviceSpecs = append(deviceSpecs, cdrom)

	serial, err := s.getSerial(datastore, vm, devices)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting serial device")
	}
	deviceSpecs = append(deviceSpecs, serial)
	// this extra config prevents the asking of a blocking question during the clone
	// see https://kb.vmware.com/s/article/1027096
	dontAskExtraConfig := []types.BaseOptionValue{}
	dontAskExtraConfig = append(dontAskExtraConfig, &types.OptionValue{
		Key:   "answer.msg.serial.file.open",
		Value: "Replace",
	})


	deviceSpecs = append(deviceSpecs, cdrom)

	spec := types.VirtualMachineCloneSpec{
		Location: types.VirtualMachineRelocateSpec{
			Datastore:    types.NewReference(datastore.Reference()),
			DiskMoveType: diskMoveType,
			Folder:       types.NewReference(folder.Reference()),
			Pool:         types.NewReference(pool.Reference()),
		},
		Config: &types.VirtualMachineConfigSpec{
			Annotation:   "Created by karina from " + vm.Template,
			Flags:        newVMFlagInfo(),
			DeviceChange: deviceSpecs,
			NumCPUs:      vm.CPUs,
			MemoryMB:     vm.MemoryGB * 1024,
			ExtraConfig: dontAskExtraConfig,
		},
		PowerOn: true,
	}

	s.Infof("Cloning %s to %s", vm.Template, vm.Name)

	task, err := tpl.Clone(ctx, folder, vm.Name, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "error trigging clone op for machine %s/%s %+v\n\n%+v", folder, vm.Name, err, spec)
	}

	_, err = task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("clone: failed create waiter: %v", err)
	}

	obj, err := s.FindVM(vm.Name)
	if err != nil {
		return nil, fmt.Errorf("clone: failed to find VM: %v", err)
	}
	s.Infof("Cloned VM: %s", obj.UUID(ctx))
	return obj, nil
}

func newVMFlagInfo() *types.VirtualMachineFlagInfo {
	diskUUIDEnabled := true
	return &types.VirtualMachineFlagInfo{
		DiskUuidEnabled: &diskUUIDEnabled,
	}
}

func (s *Session) getCdrom(datastore *object.Datastore, vm ptypes.VM, devices object.VirtualDeviceList, config *konfigadm.Config) (types.BaseVirtualDeviceConfigSpec, error) {
	op := types.VirtualDeviceConfigSpecOperationEdit
	cdrom, err := devices.FindCdrom("")
	if err != nil {
		return nil, fmt.Errorf("getCdrom: failed to find CD ROM: %v", err)
	}

	if cdrom == nil {
		ide, err := devices.FindIDEController("")
		if err != nil {
			return nil, fmt.Errorf("getCdrom: failed to find IDE controller: %v", err)
		}
		cdrom, err = devices.CreateCdrom(ide)
		if err != nil {
			return nil, fmt.Errorf("getCdrom: failed to create CD ROM: %v", err)
		}
		op = types.VirtualDeviceConfigSpecOperationAdd
	}
	s.Debugf("Creating ISO for %s", vm.Name)
	iso, err := cloudinit.CreateISO(vm.Name, config.ToCloudInit().String())
	if err != nil {
		return nil, fmt.Errorf("getCdrom: failed to create ISO: %v", err)
	}
	path := fmt.Sprintf("cloud-init/%s.iso", vm.Name)
	s.Debugf("Uploading to [%s] %s", datastore.Name(), path)
	if err = datastore.UploadFile(context.TODO(), iso, path, &soap.DefaultUpload); err != nil {
		return nil, err
	}
	s.Tracef("Uploaded to %s", path)

	//NOTE: using the datastore Name as the vm.Datastore may be "" and
	//      the datastore may have been determined from default values.
	cdrom = devices.InsertIso(cdrom, fmt.Sprintf("[%s] %s", datastore.Name(), path))
	devices.Connect(cdrom) // nolint: errcheck
	return &types.VirtualDeviceConfigSpec{
		Operation: op,
		Device:    cdrom,
	}, nil
}

// getSerial finds the first serial device (adding a new serial device if none are found), it then
// creates a blank file in the datastore as a file backing-store for it, and sets this file as its backing-store.
// The serial device is a requirement for Ubuntu image booting which can have a range of issues
// with the default configuration if a working serial device is not present.
func (s *Session) getSerial(datastore *object.Datastore, vm ptypes.VM, devices object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {
	op := types.VirtualDeviceConfigSpecOperationEdit
	serial, err := devices.FindSerialPort("")
	if err != nil || serial == nil {
		s.Debugf("No serial device found for %s, creating a new one", vm.Name)
		serial, err = devices.CreateSerialPort()
		if err != nil {
			return nil, fmt.Errorf("getSerial: failed to create a new serial device: %v", err)
		}
		op = types.VirtualDeviceConfigSpecOperationAdd
	}
	s.Debugf("Creating serial device backing file for %s", vm.Name)
	file, err := ioutil.TempFile("/tmp", "serial-backing")
	if err != nil {
		s.Errorf("Error creating local backing file for serial device %v",err)
		return nil, err
	}
	defer os.Remove(file.Name())

	path := fmt.Sprintf("serial-devices/%s.serial", vm.Name)
	if err = datastore.UploadFile(context.TODO(), file.Name(), path, &soap.DefaultUpload); err != nil {
		return nil, err
	}
	s.Tracef("Uploaded to %s", path)


	serial.Backing = &types.VirtualSerialPortFileBackingInfo{
		VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
			FileName: fmt.Sprintf("[%s] %s", datastore.Name(), path),
		},
	}

	devices.Connect(serial) // nolint: errcheck
	
	return &types.VirtualDeviceConfigSpec{
		Operation: op,
		Device:    serial,
	}, nil
}

func getDiskSpec(vm ptypes.VM, devices object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {
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

func getNetworkSpecs(s Session, vm ptypes.VM, devices object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	ctx := context.TODO()
	deviceSpecs := []types.BaseVirtualDeviceConfigSpec{}

	// Remove any existing NICs
	for _, dev := range devices.SelectByType((*types.VirtualEthernetCard)(nil)) {
		deviceSpecs = append(deviceSpecs, &types.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: types.VirtualDeviceConfigSpecOperationRemove,
		})
	}

	id := int32(-100)
	for _, net := range vm.Network {
		ref, err := s.Finder.Network(context.TODO(), net)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to find network %q", net)
		}
		backing, err := ref.EthernetCardBackingInfo(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create new ethernet card backing info for network %q on %q", net, ctx)
		}
		dev, err := object.EthernetCardTypes().CreateEthernetCard(ethCardType, backing)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create new ethernet card %q for network %q on %q", ethCardType, net, ctx)
		}

		// Get the actual NIC object. This is safe to assert without a check
		// because "object.EthernetCardTypes().CreateEthernetCard" returns a
		// "types.BaseVirtualEthernetCard" as a "types.BaseVirtualDevice".
		nic := dev.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()

		// Assign a temporary device key to ensure that a unique one will be
		// generated when the device is created.
		nic.Key = id
		id++
		deviceSpecs = append(deviceSpecs, &types.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: types.VirtualDeviceConfigSpecOperationAdd,
		})
	}
	return deviceSpecs, nil
}
