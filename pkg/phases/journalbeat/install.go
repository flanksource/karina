package journalbeat

import (
	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if !p.Journalbeat.Enabled {
		p.Infof("Skipping deployment of journalbeat, it is disabled")
		return p.DeleteSpecs(constants.PlatformSystem, "journalbeat.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "journalbeat.yaml")
}
