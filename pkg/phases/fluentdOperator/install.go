package fluentdOperator

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "kube-fluentd-operator"
)

func Deploy(p *platform.Platform) error {
	if p.FluentdOperator == nil || p.FluentdOperator.Disabled {
		log.Infof("Skipping deployment of fluentd-operator, it is disabled")
		return nil
	} else {
		log.Infof("Deploying fluentd-operator %s", p.FluentdOperator.Version)
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	if !p.FluentdOperator.DisableDefaultConfig {
		if err := p.ApplySpecs("", "kube-fluentd-default.yaml"); err != nil {
			return err
		}
	}

	c := p.FluentdOperator.Elasticsearch
	if err := p.CreateOrUpdateSecret("fluentd", Namespace, map[string][]byte{
		"FLUENT_ELASTICSEARCH_HOST":       []byte(c.URL),
		"FLUENT_ELASTICSEARCH_PORT":       []byte(c.Port),
		"FLUENT_ELASTICSEARCH_SCHEME":     []byte(c.Scheme),
		"FLUENT_ELASTICSEARCH_SSL_VERIFY": []byte(c.Verify),
		"FLUENT_ELASTICSEARCH_USER":       []byte(c.User),
		"FLUENT_ELASTICSEARCH_PASSWORD":   []byte(c.Password),
	}); err != nil {
		return err
	}

	if p.FluentdOperator.Version == "" {
		p.FluentdOperator.Version = "1.11.0"
	}

	return p.ApplySpecs("", "kube-fluentd-operator.yaml")
}
