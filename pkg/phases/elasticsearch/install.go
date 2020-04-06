package elasticsearch

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func Deploy(p *platform.Platform) error {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		log.Infof("Skipping deployment of elasticsearch, it is disabled")
		return nil
	}

	return p.ApplySpecs("", "elasticsearch.yaml")
}
