package rabbitmqoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "rabbitmq-system"
)

func Install(p *platform.Platform) error {
	if p.RabbitmqOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "rabbitmq.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "rabbitmq.yaml")
}
