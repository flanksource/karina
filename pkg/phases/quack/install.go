package quack

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

var (
	EnabledLabels = map[string]string{
		"quack.pusher.com/enabled": "true",
	}
)

func Install(platform *platform.Platform) error {
	if platform.Quack == nil || !platform.Quack.Disabled {
		log.Infof("Installing Quack")
		return platform.ApplySpecs("", "quack.yaml")
	}

	return nil
}
