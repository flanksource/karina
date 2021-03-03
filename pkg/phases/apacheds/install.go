package apacheds

import (
	"github.com/flanksource/karina/pkg/platform"
)

const MinioNamespace = "minio"
const Namespace = "ldap"

func Install(platform *platform.Platform) error {
	if platform.Ldap != nil && !platform.Ldap.Disabled && platform.Ldap.E2E.Mock {
		if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
			return err
		}
		if err := platform.ApplySpecs(Namespace, "apacheds.yaml"); err != nil {
			platform.Errorf("Error deploying apacheds: %s\n", err)
		}
	} else {
		if err := platform.DeleteSpecs(Namespace, "apacheds.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	return nil
}
