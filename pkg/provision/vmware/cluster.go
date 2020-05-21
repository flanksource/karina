package vmware

import (
	"context"
	"fmt"

	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/types"
)

type vmwareCluster struct {
	ctx        context.Context
	vsphere    types.Vsphere
	prefix     string
	session    *Session
	vmPrefixes []string
	DryRun     bool
}

// NewVMwareCluster opens a new vmware session using environment variables
func NewVMwareCluster(platform types.PlatformConfig) (types.Cluster, error) {
	cluster := vmwareCluster{
		ctx:    context.TODO(),
		prefix: platform.HostPrefix + "-" + platform.Name,
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
	cluster.vmPrefixes = []string{platform.Master.Prefix}
	for _, vm := range platform.Nodes {
		cluster.vmPrefixes = append(cluster.vmPrefixes, vm.Prefix)
	}
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
	machines := map[string]types.Machine{}

	// To list all machines for a cluster we search by each prefix combination
	// we cannot search just using the cluster prefix as it may return incorrect startsWith results
	for _, prefix := range cluster.vmPrefixes {
		list, err := cluster.GetMachinesByPrefix(prefix)
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
