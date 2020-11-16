package vmware

import (
	"context"
	"fmt"
	"net/url"

	"github.com/flanksource/karina/pkg/types"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vapi/rest"
	vtags "github.com/vmware/govmomi/vapi/tags"
)

type vmwareCluster struct {
	ctx        context.Context
	vsphere    types.Vsphere
	prefix     string
	session    *Session
	vmPrefixes map[string]types.VM
	DryRun     bool
}

// NewVMwareCluster opens a new vmware session using environment variables
func NewVMwareCluster(platform types.PlatformConfig) (types.Cluster, error) {
	cluster := vmwareCluster{
		ctx:        context.TODO(),
		vmPrefixes: make(map[string]types.VM),
		prefix:     platform.HostPrefix + "-" + platform.Name,
	}

	if platform.Vsphere == nil {
		return nil, fmt.Errorf("failed to get session from env")
	}

	session, err := GetOrCreateCachedSession(
		platform.Vsphere.Datacenter,
		platform.Vsphere.Username,
		platform.Vsphere.Password,
		platform.Vsphere.Hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from env: %v", err)
	}
	cluster.session = session
	cluster.vsphere = *platform.Vsphere
	for _, vm := range platform.Nodes {
		cluster.vmPrefixes[vm.Prefix] = vm
	}
	cluster.vmPrefixes[platform.Master.Prefix] = platform.Master
	return &cluster, nil
}

func (cluster *vmwareCluster) Clone(template types.VM, config *konfigadm.Config) (types.Machine, error) {
	LoadGovcEnvVars(cluster.vsphere, &template)
	vm, err := cluster.session.Clone(template, config)
	if err != nil {
		return nil, err
	}
	return NewVM(cluster.ctx, cluster.DryRun, vm, &template), nil
}

func (cluster *vmwareCluster) CloneTemplate(template types.VM, config *konfigadm.Config) (types.Machine, error) {
	LoadGovcEnvVars(cluster.vsphere, &template)
	vm, err := cluster.session.CloneTemplate(template, config)
	if err != nil {
		return nil, err
	}

	return NewVM(cluster.ctx, cluster.DryRun, vm, &template), nil
}

// GetVMs returns a list of all VM's associated with the cluster
func (cluster *vmwareCluster) GetMachines() (map[string]types.Machine, error) {
	machines := map[string]types.Machine{}

	// To list all machines for a cluster we search by each prefix combination
	// we cannot search just using the cluster prefix as it may return incorrect startsWith results
	for _, vm := range cluster.vmPrefixes {
		list, err := cluster.GetMachinesFor(&vm)
		if err != nil {
			return nil, err
		}
		for name, machine := range list {
			machines[name] = machine
		}
	}
	return machines, nil
}

// GetVMs returns a list of all VM's associated with the cluster
func (cluster *vmwareCluster) GetMachinesFor(vm *types.VM) (map[string]types.Machine, error) {
	var vms = make(map[string]types.Machine)
	list, err := cluster.session.Finder.VirtualMachineList(
		cluster.ctx, fmt.Sprintf("%s-%s*", cluster.prefix, vm.Prefix))
	if err != nil {
		//ignore not found error
		return nil, nil
	}
	for _, _vm := range list {
		vms[_vm.Name()] = NewVM(cluster.ctx, cluster.DryRun, _vm, vm)
	}
	return vms, nil
}

func (cluster *vmwareCluster) GetMachine(name string) (types.Machine, error) {
	machines, err := cluster.GetMachines()
	return machines[name], err
}

func (cluster *vmwareCluster) SetTags(vm types.Machine, tags map[string]string) error {
	message := "Setting tags ["
	for k, v := range tags {
		message += fmt.Sprintf("%s=%s ", k, v)
	}
	message += fmt.Sprintf("] to virtual machine %s", vm.Name())
	log.Infof(message)

	restClient := rest.NewClient(cluster.session.Client.Client)
	user := url.UserPassword(cluster.vsphere.Username, cluster.vsphere.Password)
	if err := restClient.Login(cluster.ctx, user); err != nil {
		return errors.Wrap(err, "failed to login")
	}
	manager := vtags.NewManager(restClient)

	for categoryID, tagName := range tags {
		categoryTags, err := manager.GetTagsForCategory(cluster.ctx, categoryID)
		if err != nil {
			return errors.Wrapf(err, "failed to list tags for category %s: %v", categoryID, err)
		}
		tagID := ""
		for _, t := range categoryTags {
			if t.Name == tagName {
				tagID = t.ID
			}
		}

		log.Debugf("Found tag ID: %s", tagID)

		if tagID != "" {
			return manager.AttachTag(cluster.ctx, tagID, vm.Reference())
		}
	}

	return nil
}
