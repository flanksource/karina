package vmware

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"

	"github.com/flanksource/karina/pkg/types"
)

func (s Session) GetVMs(ctx context.Context, prefix string, vm *types.VM) ([]*object.VirtualMachine, error) {
	if vm != nil && vm.Prefix != "" {
		prefix += "-" + vm.Prefix
	}
	list, err := s.Finder.VirtualMachineList(ctx, prefix+"*")
	if err != nil {
		return nil, fmt.Errorf("getCdrom: getVMs to list VMs: %v", err)
	}
	return list, nil
}
