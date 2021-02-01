package istiooperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "istio-operator"
)

func Install(p *platform.Platform) error {
	if p.IstioOperator.IsDisabled() {
		p.Logger.Errorf("DELETING because Istio operator disabled")
		return p.DeleteSpecs(Namespace, "istio-operator.yaml")
	}

	p.Logger.Errorf("APPLYING because Istio operator enabled")
	return p.ApplySpecs(Namespace, "istio-operator.yaml")
}
