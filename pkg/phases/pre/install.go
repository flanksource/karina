package pre

import (
	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/platform"
)

func Install(platform *platform.Platform) error {
	// If more preinstall functions are needed, add per function error checking
	return certmanager.PreInstall(platform)
}
