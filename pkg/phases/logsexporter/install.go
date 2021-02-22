package logsexporter

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Install(p *platform.Platform) error {
	if p.LogsExporter.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "logs-exporter.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "logs-exporter.yaml")
}
