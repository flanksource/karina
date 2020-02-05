package opa

import (
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
		return err
	}
	kubectl := platform.GetKubectl()
	for index := range platform.OPA.NamespaceWhitelist {
		err := kubectl("get ns %s &> /dev/null", platform.OPA.NamespaceWhitelist[index])
		if err == nil {
			kubectl("label ns %s openpolicyagent.org/webhook=ignore --overwrite &> /dev/null", platform.OPA.NamespaceWhitelist[index])
		}
	}

	return platform.ApplySpecs(Namespace, "opa.yaml")
}
