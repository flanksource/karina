package apacheds

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "ldap"

func Install(platform *platform.Platform) error {
	if platform.Ldap == nil || platform.Ldap.Disabled || !platform.Ldap.E2E.Mock {
		return platform.DeleteSpecs(Namespace, "apacheds.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "apacheds.yaml")
}
