package eck

import (
	"fmt"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const Namespace = "elastic-system"

func Deploy(p *platform.Platform) error {
	if p.ECK == nil || p.ECK.Disabled {
		p.Infof("Skipping deployment of ECK, it is disabled")
		return nil
	}
	p.Infof("Deploying ECK %s", p.ECK.Version)
	if err := p.CreateOrUpdateNamespace(Namespace, nil, p.DefaultNamespaceAnnotations()); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "eck.yaml")
}
