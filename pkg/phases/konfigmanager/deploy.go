package konfigmanager

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = constants.PlatformSystem
	Name      = "konfig-manager"
)

// Deploy deploys the konfig-manager into the platform-system namespace
func Deploy(p *platform.Platform) error {
	if p.KonfigManager.Version == "" {
		p.KonfigManager.Version = "v0.2.1"
	}
	if p.KonfigManager.IsDisabled() {
		return p.DeleteSpecs(Namespace, "konfig-manager.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}
	return p.ApplySpecs(Namespace, "konfig-manager.yaml")
}
