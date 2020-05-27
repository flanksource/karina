package packetbeat

import (
	"github.com/pkg/errors"

	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Packetbeat.Disabled {
		p.Infof("Skipping deployment of packetbeat, it is disabled")
		return nil
	}

	if p.Packetbeat.Elasticsearch != nil {
		err := p.GetOrCreateSecret("elastic-packetbeat", constants.PlatformSystem, map[string][]byte{
			"ELASTIC_URL":      []byte(p.Packetbeat.Elasticsearch.GetURL()),
			"ELASTIC_USERNAME": []byte(p.Packetbeat.Elasticsearch.User),
			"ELASTIC_PASSWORD": []byte(p.Packetbeat.Elasticsearch.Password),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create secret elastic")
		}
	}

	return p.ApplySpecs(constants.PlatformSystem, "packetbeat.yaml")
}
