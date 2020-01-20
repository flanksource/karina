package stubs

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if err := platform.ApplySpecs("", "minio.yaml"); err != nil {
		log.Errorf("Error deploying minio: %s\n", err)
	}
	return nil
}
