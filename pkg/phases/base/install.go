package base

import (
	"os"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Install(platform *platform.Platform) error {
	os.Mkdir(".bin", 0755) // nolint: errcheck

	if err := platform.ApplySpecs("", "rbac.yaml"); err != nil {
		platform.Errorf("Error deploying base rbac: %s", err)
	}

	if err := platform.CreateOrUpdateNamespace(constants.KubeSystem, nil, nil); err != nil {
		platform.Errorf("Error deploying base kube-system labels/annotations: %s", err)
	}

	if err := platform.CreateOrUpdateNamespace("monitoring", nil, nil); err != nil {
		platform.Errorf("Error deploying base monitoring labels/annotations: %s", err)
	}

	if err := platform.CreateOrUpdateNamespace(constants.PlatformSystem, map[string]string{"quack.pusher.com/enabled": "true"}, nil); err != nil {
		platform.Errorf("Error deploying base platform-system labels/annotations: %s", err)
	}

	return nil
}
