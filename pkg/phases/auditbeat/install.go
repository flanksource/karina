package auditbeat

import (
	"github.com/pkg/errors"

	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Auditbeat.Disabled {
		p.Infof("Skipping deployment of auditbeat, it is disabled")
		return nil
	}

	if p.Auditbeat.Elasticsearch != nil {
		err := p.GetOrCreateSecret("elastic-auditbeat", constants.PlatformSystem, map[string][]byte{
			"ELASTIC_URL":      []byte(p.Auditbeat.Elasticsearch.GetURL()),
			"ELASTIC_USERNAME": []byte(p.Auditbeat.Elasticsearch.User),
			"ELASTIC_PASSWORD": []byte(p.Auditbeat.Elasticsearch.Password),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create secret elastic-auditbeat")
		}
	}

	return p.ApplySpecs(constants.PlatformSystem, "auditbeat.yaml")
}
