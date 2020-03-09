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

	if platform.OPA.LogLevel == "" {
		platform.OPA.LogLevel = "error"
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, map[string]string{
		"app": "opa",
	}, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	for index := range platform.OPA.NamespaceWhitelist {
		err := platform.CreateOrUpdateNamespace(platform.OPA.NamespaceWhitelist[index], nil, nil)
		if err != nil {
			fmt.Println(err)
		}
	}

	return platform.ApplySpecs(Namespace, "opa.yaml")
}
