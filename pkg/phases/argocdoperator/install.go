package argocdoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "argocd"
)

func Deploy(platform *platform.Platform) error {
	if platform.ArgoCDOperator.IsDisabled() {
		if err := platform.DeleteSpecs("", "argocd-operator.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if platform.ArgoCDOperator.Version == "" {
		platform.ArgoCDOperator.Version = "v0.0.15"
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "argocd-operator.yaml")
}
