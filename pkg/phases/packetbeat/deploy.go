package packetbeat

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Packetbeat.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "packetbeat.yaml")
	}
	return p.ApplySpecs(constants.PlatformSystem, "packetbeat.yaml")
}
