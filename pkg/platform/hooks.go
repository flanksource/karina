package platform

import "github.com/flanksource/karina/pkg/types"

type ProvisionHook interface {
	BeforeProvision(platform *Platform, machine types.VM) error
	AfterProvision(platform *Platform, machine types.Machine) error
	BeforeTerminate(platform *Platform, machine types.Machine) error
	AfterTerminate(platform *Platform, machine types.Machine) error
}
type LoadBalancerProvider interface {
	ProvisionHook
	GetControlPlaneEndpoint(platform *Platform) (string, error)
}

type CNIProvider interface {
	ProvisionHook
}
