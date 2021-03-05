package antrea

import (
	"time"

	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace  = "kube-system"
	Controller = "antrea-controller"
	DaemonSet  = "antrea-agent"
)

func Install(p *platform.Platform) error {
	if !p.Calico.IsDisabled() && !p.Antrea.IsDisabled() {
		p.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if p.Antrea.IsDisabled() {
		return p.DeleteSpecs(Namespace, "antrea.yaml")
	}
	p.Antrea.IsCertReady = false
	if err := p.ApplySpecs(Namespace, "antrea.yaml"); err != nil {
		return err
	}

	if err := p.WaitForDeployment(Namespace, Controller, 10*time.Minute); err != nil {
		return err
	}
	return p.WaitForDaemonSet(Namespace, DaemonSet, 10*time.Minute)
}
