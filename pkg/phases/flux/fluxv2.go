package flux

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "flux-system"

func InstallV2(p *platform.Platform) error {
	if !p.Flux.Enabled {
		return p.DeleteSpecs(Namespace, "flux-v2.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "flux-v2.yaml")
}
