package elasticsearch

import (
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const Namespace = "eck"

func Deploy(p *platform.Platform) error {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}
	if p.Elasticsearch.Mem == nil {
		p.Elasticsearch.Mem = &types.Memory{Limits: "3G", Requests: "2G"}
	}
	if p.Elasticsearch.Replicas == 0 {
		p.Elasticsearch.Replicas = 3
	}
	if p.OAuth2Proxy == nil {
		p.OAuth2Proxy = &types.OAuth2Proxy{
			Disabled: true,
		}
	}

	if p.Elasticsearch.Persistence == nil {
		p.Elasticsearch.Persistence = &types.Persistence{
			XEnabled: types.XEnabled{
				Disabled: true,
			},
			StorageClass: "",
			Capacity:     "",
		}
	}

	return p.ApplySpecs(Namespace, "elasticsearch.yaml")
}
