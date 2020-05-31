package auditbeat

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if !p.Auditbeat.Enabled {
		p.Infof("Skipping deployment of auditbeat, it is disabled")
		return p.DeleteSpecs(constants.PlatformSystem, "auditbeat.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "auditbeat.yaml")
}
