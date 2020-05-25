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
		return platform.ApplySpecs(Namespace, "quack.yaml")
	}

	if err := platform.DeleteSpecs(Namespace, "quack.yaml"); err != nil {
		platform.Warnf("failed to delete specs: %v", err)
	}

	return nil
}
