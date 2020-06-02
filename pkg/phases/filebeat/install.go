package filebeat

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	for _, f := range p.Filebeat {
		if f.Disabled {
			p.Infof("Skipping deployment of filebeat %s, it is disabled", f.Name)
			continue
		}

		if f.Elasticsearch != nil {
			secretName := fmt.Sprintf("elastic-%s", f.Name)
			err := p.GetOrCreateSecret(secretName, constants.PlatformSystem, map[string][]byte{
				"ELASTIC_URL":      []byte(f.Elasticsearch.GetURL()),
				"ELASTIC_USERNAME": []byte(f.Elasticsearch.User),
				"ELASTIC_PASSWORD": []byte(f.Elasticsearch.Password),
			})
			if err != nil {
				return errors.Wrap(err, "failed to create secret elastic")
			}
		}

		if f.Logstash != nil {
			secretName := fmt.Sprintf("logstash-%s", f.Name)
			err := p.GetOrCreateSecret(secretName, constants.PlatformSystem, map[string][]byte{
				"LOGSTASH_URL":      []byte(f.Logstash.GetURL()),
				"LOGSTASH_USERNAME": []byte(f.Logstash.User),
				"LOGSTASH_PASSWORD": []byte(f.Logstash.Password),
			})
			if err != nil {
				return errors.Wrap(err, "Failed to create secret logstash")
			}
		}
	}

	return p.ApplySpecs(constants.PlatformSystem, "filebeat.yaml")
}
