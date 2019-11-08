package vmware

import (
	"context"

	"github.com/vmware/govmomi/object"

	"github.com/moshloop/platform-cli/pkg/types"
)

func (s Session) GetVMs(ctx context.Context, prefix string, vm *types.VM) ([]*object.VirtualMachine, error) {
	if vm != nil && vm.Prefix != "" {
		prefix += "-" + vm.Prefix
	}
	list, err := s.Finder.VirtualMachineList(ctx, prefix+"*")
	if err != nil {
		return nil, err
	}
	return list, nil
}
