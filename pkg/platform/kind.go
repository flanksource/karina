package platform

import (
	"github.com/flanksource/karina/pkg/types"
)

type KindProvider struct {
}

func (kind KindProvider) BeforeProvision(platform *Platform, machine *types.VM) error {
	return nil
}

func (kind KindProvider) AfterProvision(platform *Platform, machine types.Machine) error {
	return nil
}

func (kind KindProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func (kind KindProvider) AfterTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func (kind KindProvider) GetControlPlaneEndpoint(platform *Platform) (string, error) {
	return "localhost:8443", nil
}

func (kind KindProvider) GetExternalEndpoints(platform *Platform) ([]string, error) {
	return []string{"localhost"}, nil
}

func (kind KindProvider) String() string {
	return "kind"
}
