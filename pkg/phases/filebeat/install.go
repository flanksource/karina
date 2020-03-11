package filebeat

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Filebeat == nil || p.Filebeat.Disabled {
		log.Infof("Skipping deployment of filebeat, it is disabled")
		return nil
	}

	if p.Filebeat.Elasticsearch != nil {
		p.GetOrCreateSecret("elastic", constants.PlatformSystem, map[string][]byte{
			"ELASTIC_URL":      []byte(p.Filebeat.Elasticsearch.GetURL()),
			"ELASTIC_USERNAME": []byte(p.Filebeat.Elasticsearch.User),
			"ELASTIC_PASSWORD": []byte(p.Filebeat.Elasticsearch.Password),
		})
	}

	if p.Filebeat.Logstash != nil {
		p.GetOrCreateSecret("logstash", constants.PlatformSystem, map[string][]byte{
			"LOGSTASH_URL":      []byte(p.Filebeat.Logstash.GetURL()),
			"LOGSTASH_USERNAME": []byte(p.Filebeat.Logstash.User),
			"LOGSTASH_PASSWORD": []byte(p.Filebeat.Logstash.Password),
		})
	}

	return p.ApplySpecs(constants.PlatformSystem, "filebeat.yaml")
}
