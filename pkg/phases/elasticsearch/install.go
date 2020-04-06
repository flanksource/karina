package elasticsearch

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

const Namespace = "eck"

func Deploy(p *platform.Platform) error {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		log.Infof("Skipping deployment of elasticsearch, it is disabled")
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "elasticsearch.yaml")
}
