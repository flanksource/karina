package gitoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "platform-system"
)

func Install(p *platform.Platform) error {
	if p.GitOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "git-operator.yaml")
	}

	return p.ApplySpecs(Namespace, "git-operator.yaml")
}
