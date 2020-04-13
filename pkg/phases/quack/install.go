package quack

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

var (
	EnabledLabels = map[string]string{
		"quack.pusher.com/enabled": "true",
	}
)

const Namespace = "quack"

func Install(platform *platform.Platform) error {
	if platform.Quack == nil || !platform.Quack.Disabled {
		platform.Infof("Installing Quack")
		return platform.ApplySpecs("", "quack.yaml")
	}
	return nil
}
