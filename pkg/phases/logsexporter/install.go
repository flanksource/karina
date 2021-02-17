package logsexporter

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
)

func Install(p *platform.Platform) error {
	if p.LogsExporter.IsDisabled() {
		return p.DeleteSpecs(constants.PlatformSystem, "logs-exporter.yaml")
	}

	data := map[string][]byte{
		"username": []byte(p.LogsExporter.Elasticsearch.User),
		"password": []byte(p.LogsExporter.Elasticsearch.Password),
		"url":      []byte(p.LogsExporter.Elasticsearch.URL),
	}
	if err := p.GetOrCreateSecret("logs-exporter-elastic", constants.PlatformSystem, data); err != nil {
		return errors.Wrap(err, "failed to create secret logs-exporter-elastic")
	}

	return p.ApplySpecs(constants.PlatformSystem, "logs-exporter.yaml")
}
