package localpath

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "local-path-storage"

func Install(platform *platform.Platform) error {
	if platform.LocalPath.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "local-path.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}
	return platform.ApplySpecs(Namespace, "local-path.yaml")
}
