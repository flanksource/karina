package vmware

import (
	"context"
	"fmt"

	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/types"
)

type vmwareCluster struct {
	ctx     context.Context
	vsphere types.Vsphere
	prefix  string
	session *Session
	DryRun  bool
}

// NewVMwareCluster opens a new vmware session using environment variables
func NewVMwareCluster(platform types.PlatformConfig) (types.Cluster, error) {
	cluster := vmwareCluster{
		ctx:    context.TODO(),
		prefix: platform.HostPrefix + "-" + platform.Name,
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
	return &cluster, nil
}

func (cluster *vmwareCluster) Clone(template types.VM, config *konfigadm.Config) (types.Machine, error) {
	LoadGovcEnvVars(cluster.vsphere, &template)
	vm, err := cluster.session.Clone(template, config)
	if err != nil {
		return nil, err
	}
	return NewVM(cluster.ctx, cluster.DryRun, vm), nil
}

// GetVMs returns a list of all VM's associated with the cluster
func (cluster *vmwareCluster) GetMachines() (map[string]types.Machine, error) {
	return cluster.GetMachinesByPrefix("")
}

// GetVMs returns a list of all VM's associated with the cluster
func (cluster *vmwareCluster) GetMachinesByPrefix(prefix string) (map[string]types.Machine, error) {
	var vms = make(map[string]types.Machine)
	list, err := cluster.session.Finder.VirtualMachineList(
		cluster.ctx, fmt.Sprintf("%s-%s*", cluster.prefix, prefix))
	if err != nil {
		//ignore not found error
		return nil, nil
	}
	for _, vm := range list {
		vms[vm.Name()] = NewVM(cluster.ctx, cluster.DryRun, vm)
	}
	return vms, nil
}

func (cluster *vmwareCluster) GetMachine(name string) (types.Machine, error) {
	machines, err := cluster.GetMachines()
	return machines[name], err
}
