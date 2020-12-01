package antrea

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(p *platform.Platform) error {
	if (p.Calico != nil && !p.Calico.IsDisabled()) && (p.Antrea != nil && !p.Antrea.IsDisabled()) {
		p.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if p.Antrea == nil || p.Antrea.IsDisabled() {
		if err := p.DeleteSpecs(Namespace, "antrea.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	p.Antrea.IsCertReady = false
	return p.ApplySpecs(Namespace, "antrea.yaml")
}
