package configmapreloader

import (
	"github.com/flanksource/commons/utils"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.ConfigMapReloader.Disabled {
		if err := p.DeleteSpecs(constants.PlatformSystem, "configmap-reloader.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if p.ConfigMapReloader.Version == "" {
		p.ConfigMapReloader.Version = "v0.0.56"
	} else {
		p.ConfigMapReloader.Version = utils.NormalizeVersion(p.ConfigMapReloader.Version)
	}

	if err := p.CreateOrUpdateNamespace(constants.PlatformSystem, nil, nil); err != nil {
		return err
	}
	return p.ApplySpecs("", "configmap-reloader.yaml")
}
