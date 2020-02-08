package opa

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

const (
	Namespace = "opa"
)

func Install(platform *platform.Platform) error {
	if platform.OPA == nil || platform.OPA.Disabled {
		return nil
	}
	if platform.OPA.KubeMgmtVersion == "" {
		platform.OPA.KubeMgmtVersion = "0.8"
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, map[string]string{
		"app": "opa",
	}, nil); err != nil {
		log.Tracef("Install: Failed to create/update namespace: %s", err)
		return err
	}
	return platform.ApplySpecs(Namespace, "opa.yaml")
}
