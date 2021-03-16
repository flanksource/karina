package calico

import (
	"time"

	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace  = "kube-system"
	Controller = "calico-kube-controllers"
	DaemonSet  = "calico-node"
)

func Install(platform *platform.Platform) error {
	if !platform.Calico.IsDisabled() && !platform.Antrea.IsDisabled() {
		platform.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if platform.Calico.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "calico.yaml")
	}

	if err := platform.ApplySpecs(Namespace, "calico.yaml"); err != nil {
		return err
	}

	if err := platform.WaitForDeployment(Namespace, Controller, 2*time.Minute); err != nil {
		return err
	}
	return platform.WaitForDaemonSet(Namespace, DaemonSet, 2*time.Minute)
}
