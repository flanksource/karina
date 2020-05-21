package configmapreloader

import (
	"github.com/flanksource/commons/utils"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "platform-system"
)

func Deploy(p *platform.Platform) error {
	if p.ConfigMapReloader.Disabled {
		if err := p.DeleteSpecs(Namespace, "configmap-reloader.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if p.ConfigMapReloader.Version == "" {
		p.ConfigMapReloader.Version = "v0.0.56"
	} else {
		p.ConfigMapReloader.Version = utils.NormalizeVersion(p.ConfigMapReloader.Version)
	}

	p.Infof("Deploying configmap-reloader %s into %s", p.ConfigMapReloader.Version, Namespace)

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs("", "configmap-reloader.yaml")
}
