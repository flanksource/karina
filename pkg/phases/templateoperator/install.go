package templateoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "platform-system"
)

func Install(p *platform.Platform) error {
	if p.TemplateOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "template-operator.yaml")
	}

	return p.ApplySpecs(Namespace, "template-operator.yaml")
}
