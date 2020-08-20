package calico

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(platform *platform.Platform) error {
	if platform.Calico == nil || platform.Calico.IsDisabled() {
		if err := platform.DeleteSpecs(Namespace, "calico.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	return platform.ApplySpecs(Namespace, "calico.yaml")
}
