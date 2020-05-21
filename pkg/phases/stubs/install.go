package stubs

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if platform.S3.E2E.Minio {
		if err := platform.CreateOrUpdateNamespace("minio", nil, nil); err != nil {
			return err
		}
		platform.Infof("Installing minio")
		if err := platform.ApplySpecs("", "minio.yaml"); err != nil {
			platform.Errorf("Error deploying minio: %s\n", err)
		}
	} else {
		if err := platform.DeleteSpecs("", "minio.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	if platform.Ldap != nil && platform.Ldap.E2E.Mock {
		if err := platform.CreateOrUpdateNamespace("ldap", nil, nil); err != nil {
			return err
		}
		if err := platform.ApplySpecs("", "apacheds.yaml"); err != nil {
			platform.Errorf("Error deploying apacheds: %s\n", err)
		}
	} else {
		if err := platform.DeleteSpecs("", "apacheds.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	return nil
}
