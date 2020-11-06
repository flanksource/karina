package redisoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "redis-operator"
)

func Install(p *platform.Platform) error {
	if p.RedisOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "redis-operator.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "redis-operator.yaml")
}
