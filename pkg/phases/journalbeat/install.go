package journalbeat

import (
	"github.com/pkg/errors"

	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Journalbeat.Disabled {
		p.Infof("Skipping deployment of journalbeat, it is disabled")
		return nil
	}

	if p.Journalbeat.Elasticsearch != nil {
		err := p.GetOrCreateSecret("elastic-journalbeat", constants.PlatformSystem, map[string][]byte{
			"ELASTIC_URL":      []byte(p.Journalbeat.Elasticsearch.GetURL()),
			"ELASTIC_USERNAME": []byte(p.Journalbeat.Elasticsearch.User),
			"ELASTIC_PASSWORD": []byte(p.Journalbeat.Elasticsearch.Password),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create secret elastic")
		}
	}

	return p.ApplySpecs(constants.PlatformSystem, "journalbeat.yaml")
}
