package argorollouts

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "argo-rollouts"
)

func Deploy(platform *platform.Platform) error {
	if platform.ArgoRollouts.IsDisabled() {
		return platform.DeleteSpecs("", "argo-rollouts.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "argo-rollouts.yaml")
}
