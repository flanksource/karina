package vmware

import (
	"context"
	"fmt"

	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/types"
)

type vmwareCluster struct {
	ctx     context.Context
	prefix  string
	session *Session
	DryRun  bool
}

// NewVMwareCluster opens a new vmware session using environment variables
func NewVMwareCluster(prefix string) (types.Cluster, error) {
	cluster := vmwareCluster{
		ctx:    context.TODO(),
		prefix: prefix,
	}
	session, err := GetSessionFromEnv()
	if err != nil {
		return nil, fmt.Errorf("openViaEnv: failed to get session from env: %v", err)
	}
	cluster.session = session
	return &cluster, nil
}

func (p *vmwareCluster) Clone(template types.VM, config *konfigadm.Config) (types.Machine, error) {
	LoadGovcEnvVars(&template)
	vm, err := p.session.Clone(template, config)
	if err != nil {
		return nil, err
	}
	return NewVM(p.ctx, p.DryRun, vm), nil
}

// GetVMs returns a list of all VM's associated with the cluster
func (p *vmwareCluster) GetMachines() (map[string]types.Machine, error) {
	return p.GetMachinesByPrefix("")
}

// GetVMs returns a list of all VM's associated with the cluster
func (p *vmwareCluster) GetMachinesByPrefix(prefix string) (map[string]types.Machine, error) {
	var vms = make(map[string]types.Machine)
	list, err := p.session.Finder.VirtualMachineList(
		p.ctx, fmt.Sprintf("%s-%s*", p.prefix, prefix))
	if err != nil {
		//ignore not found error
		return nil, nil
	}
	for _, vm := range list {
		vms[vm.Name()] = NewVM(p.ctx, p.DryRun, vm)
	}
	return vms, nil
}

func (p *vmwareCluster) GetMachine(name string) (types.Machine, error) {
	machines, err := p.GetMachines()
	return machines[name], err
}
