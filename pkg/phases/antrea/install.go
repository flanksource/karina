package antrea

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(p *platform.Platform) error {
	if p.Antrea == nil || p.Antrea.IsDisabled() {
		if err := p.DeleteSpecs(Namespace, "antrea.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	return p.ApplySpecs(Namespace, "antrea.yaml")
}
