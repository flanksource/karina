package provision

import (
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

type ProvisionHook interface {
	BeforeProvision(platform *platform.Platform, vm *types.Machine) error
	AfterProvision(platform *platform.Platform, vm *types.Machine) error
	BeforeTerminate(platform *platform.Platform, vm *types.Machine) error
	AfterTerminate(platform *platform.Platform, vm *types.Machine) error
}
