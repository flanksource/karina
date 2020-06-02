package platform

import (
	"fmt"

	"github.com/flanksource/karina/pkg/types"
)

type ProvisionHook interface {
	fmt.Stringer
	BeforeProvision(platform *Platform, machine *types.VM) error
	AfterProvision(platform *Platform, machine types.Machine) error
	BeforeTerminate(platform *Platform, machine types.Machine) error
	AfterTerminate(platform *Platform, machine types.Machine) error
}

type CompositeHook struct {
	Hooks []ProvisionHook
}

func (c CompositeHook) BeforeProvision(platform *Platform, machine *types.VM) error {
	for _, hook := range c.Hooks {
		if err := hook.BeforeProvision(platform, machine); err != nil {
			return err
		}
	}
	return nil
}
func (c CompositeHook) AfterProvision(platform *Platform, machine types.Machine) error {
	var err error
	for _, hook := range c.Hooks {
		if _err := hook.AfterProvision(platform, machine); _err != nil {
			err = _err
		}
	}
	return err
}
func (c CompositeHook) BeforeTerminate(platform *Platform, machine types.Machine) error {
	for _, hook := range c.Hooks {
		if err := hook.BeforeTerminate(platform, machine); err != nil {
			return err
		}
	}
	return nil
}
func (c CompositeHook) AfterTerminate(platform *Platform, machine types.Machine) error {
	var err error
	for _, hook := range c.Hooks {
		if _err := hook.AfterTerminate(platform, machine); _err != nil {
			err = _err
		}
	}
	return err
}

func (c CompositeHook) String() string {
	return fmt.Sprintf("%v", c.Hooks)
}

type MasterDiscovery interface {
	fmt.Stringer
	GetControlPlaneEndpoint(platform *Platform) (string, error)
	GetExternalEndpoints(platform *Platform) ([]string, error)
}
