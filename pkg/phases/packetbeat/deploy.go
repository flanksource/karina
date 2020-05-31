package packetbeat

import (
	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if !p.Packetbeat.Enabled {
		p.Infof("Skipping deployment of packetbeat, it is disabled")
		return p.DeleteSpecs(constants.PlatformSystem, "packetbeat.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "packetbeat.yaml")
}
