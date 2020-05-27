package eventrouter

import (
	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.EventRouter.Disabled {
		p.Infof("Skipping deployment of eventrouter, it is disabled")
		return nil
	}

	return p.ApplySpecs(constants.PlatformSystem, "eventrouter.yaml")
}
