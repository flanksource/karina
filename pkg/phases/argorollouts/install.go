package argorollouts

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(platform *platform.Platform) error {
	if platform.ArgoRollouts.IsDisabled() {
		return platform.DeleteSpecs("", "argo-rollouts.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(constants.PlatformSystem, nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(constants.PlatformSystem, "argo-rollouts.yaml")
}
