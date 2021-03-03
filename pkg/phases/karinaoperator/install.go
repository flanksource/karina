package karinaoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "platform-system"
)

func Install(p *platform.Platform) error {
	if p.KarinaOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "karina-operator.yaml")
	}

	return p.ApplySpecs(Namespace, "karina-operator.yaml")
}
