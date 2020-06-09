package canary

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

// Deploy deploys the canary-checker into the Platform System namespace
func Deploy(p *platform.Platform) error {
	if !p.Canary.Enabled {
		p.Infof("Skipping deployment of canary-checker, it is disabled")
		return nil
	}

	return p.ApplySpecs(constants.PlatformSystem, "canary.yaml")
}
