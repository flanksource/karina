package auditbeat

import (
	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if !p.Auditbeat.Enabled {
		p.Infof("Skipping deployment of auditbeat, it is disabled")
		return p.DeleteSpecs(constants.PlatformSystem, "auditbeat.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "auditbeat.yaml")
}
