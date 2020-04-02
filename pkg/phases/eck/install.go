package eck

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const Namespace = "elastic-system"

func Deploy(p *platform.Platform) error {
	if p.ECK == nil || p.ECK.Disabled {
		log.Infof("Skipping deployment of ECK, it is disabled")
		return nil
	}
	log.Infof("Deploying ECK %s", p.ECK.Version)
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "eck.yaml")
}
