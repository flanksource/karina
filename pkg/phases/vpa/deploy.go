package vpa

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = constants.KubeSystem
)

func Install(p *platform.Platform) error {
	if p.VPA == nil || p.VPA.Version == "" {
		return p.DeleteSpecs(Namespace, "vpa.yaml")
	}

	return p.ApplySpecs(Namespace, "vpa.yaml")
}
