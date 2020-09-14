package journalbeat

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Journalbeat.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "journalbeat.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "journalbeat.yaml")
}
