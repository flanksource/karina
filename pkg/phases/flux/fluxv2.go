package flux

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "flux-system"

func InstallV2(p *platform.Platform) error {
	if p.Flux == nil || !p.Flux.Enabled {
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "flux.yaml")
}
