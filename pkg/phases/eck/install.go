package eck

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const Namespace = "elastic-system"

func Deploy(p *platform.Platform) error {
	if p.ECK == nil || p.ECK.Disabled {
		log.Infof("Skipping deployment of ECK, it is disabled")
		return nil
	} else {
		log.Infof("Deploying ECK %s", p.ECK.Version)
	}

	return p.ApplySpecs(Namespace, "eck.yml")
}

func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return strings.TrimSuffix(version, "v")
	}
	return version
}
