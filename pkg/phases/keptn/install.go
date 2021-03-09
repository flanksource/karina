package keptn

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "keptn"
)

func Deploy(platform *platform.Platform) error {
	if platform.Keptn.IsDisabled() {
		if err := platform.DeleteSpecs(Namespace, "keptn.yaml"); err != nil {
			return err
		}
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	// Trim the first character e.g. v0.7.3 -> 0.7.3
	platform.Keptn.Version = platform.Keptn.Version[1:]

	return platform.ApplySpecs(Namespace, "keptn.yaml")
}
