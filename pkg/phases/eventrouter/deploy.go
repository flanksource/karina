package eventrouter

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.EventRouter.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "eventrouter.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "eventrouter.yaml")
}
