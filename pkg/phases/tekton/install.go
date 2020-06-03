package tekton

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "tekton-pipelines"
)

func Install(p *platform.Platform) error {
	if p.Tekton.Disabled || p.Tekton.Version == "" {
		return p.DeleteSpecs(Namespace, "tekton.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "tekton.yaml")
}
