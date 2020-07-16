package kiosk

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Kiosk.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "kiosk.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "kiosk.yaml")
}
