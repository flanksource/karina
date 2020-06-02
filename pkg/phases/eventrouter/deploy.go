package eventrouter

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.EventRouter.IsDisabled() {
		p.Infof("Skipping deployment of eventrouter, it is disabled")
		return nil
	}

	return p.ApplySpecs(constants.PlatformSystem, "eventrouter.yaml")
}
