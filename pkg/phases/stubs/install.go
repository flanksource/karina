package stubs

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if platform.Minio == nil || !platform.Minio.Disabled {
		if err := platform.CreateOrUpdateNamespace("minio", nil, nil); err != nil {
			return err
		}
		platform.Infof("Installing minio")
		if err := platform.ApplySpecs("", "minio.yaml"); err != nil {
			platform.Errorf("Error deploying minio: %s\n", err)
		}
	}
	if err := platform.CreateOrUpdateNamespace("ldap", nil, nil); err != nil {
		return err
	}
	if err := platform.ApplySpecs("", "apacheds.yaml"); err != nil {
		platform.Errorf("Error deploying apacheds: %s\n", err)
	}

	return nil
}
