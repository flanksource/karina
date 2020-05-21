package quack

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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
		if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
			return errors.Wrap(err, "failed to create/update namespace quack")
		}
		// quack gets deployed across both quack ane kube-system namespaces
		return platform.ApplySpecs(v1.NamespaceAll, "quack.yaml")
	}

	if err := platform.DeleteSpecs(v1.NamespaceAll, "quack.yaml"); err != nil {
		platform.Warnf("failed to delete specs: %v", err)
	}

	return nil
}
