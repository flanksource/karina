package istio

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "istio-operator"
)

func Install(p *platform.Platform) error {
	if p.IstioOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "istio-operator.yaml")
	}

	return p.ApplySpecs(Namespace, "istio-operator.yaml")
}
