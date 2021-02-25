package antrea

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(p *platform.Platform) error {
	if !p.Calico.IsDisabled() && !p.Antrea.IsDisabled() {
		p.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if p.Antrea.IsDisabled() {
		return p.DeleteSpecs(Namespace, "antrea.yaml")
	}
	p.Antrea.IsCertReady = false
	return p.ApplySpecs(Namespace, "antrea.yaml")
}
