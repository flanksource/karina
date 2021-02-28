package templateoperator

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = constants.PlatformSystem

func Install(p *platform.Platform) error {
	if p.TemplateOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "template-operator.yaml")
	}

	if err := p.CreateOrUpdateNamespace(constants.PlatformSystem, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "template-operator.yaml")
}
