package crds

import (
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
)

func Install(p *platform.Platform) error {
	if err := p.ApplySpecs("", "cert-manager-crd.yaml"); err != nil {
		return errors.Wrap(err, "failed to deploy cert manager CRD")
	}
	return p.ApplySpecs("", "monitoring/prometheus-crd.yaml")
}
