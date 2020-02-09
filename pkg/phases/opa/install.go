package opa

import (
	"fmt"
	"github.com/moshloop/platform-cli/pkg/platform"
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
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}
	return platform.ApplySpecs(Namespace, "opa.yaml")
}
