package elasticsearch

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

const Namespace = "eck"

func Deploy(p *platform.Platform) error {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		p.Infof("Skipping deployment of elasticsearch, it is disabled")
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}
	if p.Elasticsearch.Mem == nil {
		p.Elasticsearch.Mem = &types.Memory{Limits: "2G", Requests: "2G"}
	}
	if p.Elasticsearch.Replicas == 0 {
		p.Elasticsearch.Replicas = 3
	}

	return p.ApplySpecs(Namespace, "elasticsearch.yaml")
}
