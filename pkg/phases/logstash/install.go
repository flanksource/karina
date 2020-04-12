package logstash

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Logstash == nil || p.Logstash.Disabled {
		log.Infof("Skipping deployment of logstash, it is disabled")
		return nil
	}

	err := p.CreateOrUpdateNamespace("logstash", make(map[string]string), make(map[string]string))
	if err != nil {
		return errors.Wrap(err, "failed to create namespace logstash")
	}
	fmt.Println(p.Logstash.Elasticsearch.GetURL())
	if p.Logstash.Elasticsearch != nil {
		err := p.GetOrCreateSecret("elastic", "logstash", map[string][]byte{
			"ELASTICSEARCH_HOST":      []byte(p.Logstash.Elasticsearch.GetURL()),
			"ELASTICSEARCH_USER": []byte(p.Logstash.Elasticsearch.User),
			"ELASTICSEARCH_PASSWORD": []byte(p.Logstash.Elasticsearch.Password),
		})
		if err != nil {
			return errors.Wrap(err, "Failed to create secret elastic")
		}
	}

	return p.ApplySpecs("logstash", "logstash.yaml")
}
