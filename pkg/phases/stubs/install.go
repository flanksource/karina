package stubs

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if platform.Minio == nil || !platform.Minio.Disabled {
		log.Infof("Installing minio")
		if err := platform.ApplySpecs("", "minio.yaml"); err != nil {
			log.Errorf("Error deploying minio: %s\n", err)
		}
	}

	if err := platform.ApplySpecs("", "apacheds.yaml"); err != nil {
		log.Errorf("Error deploying apacheds: %s\n", err)
	}

	return nil
}
