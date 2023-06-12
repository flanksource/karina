package eck

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "elastic-system"

func Deploy(p *platform.Platform) error {
	if p.ECK.IsDisabled() {
		return nil
	}
	p.Infof("Deploying ECK %s", p.ECK.Version)
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "eck.yaml")
}
