package calico

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(platform *platform.Platform) error {
	if  !platform.Calico.IsDisabled()) && !platform.Antrea.IsDisabled() {
		platform.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if platform.Calico.IsDisabled() {
		return latform.DeleteSpecs(Namespace, "calico.yaml")
	}

	return platform.ApplySpecs(Namespace, "calico.yaml")
}
