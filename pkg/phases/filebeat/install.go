package filebeat

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if len(p.Filebeat) == 0 {
		return nil
	}
	for _, f := range p.Filebeat {
		if f.IsDisabled() {
			continue
		}

		if f.Elasticsearch != nil {
			var password string
			var err error
			_, password, err = p.GetEnvValue(f.Elasticsearch.Password, "eck")
			if err != nil {
				_, password, err = p.GetEnvValue(f.Elasticsearch.Password, metav1.NamespaceAll)
				if err != nil {
					return fmt.Errorf("unable to retrieve elasticsearch password for %s", f.Name)
				}
			}
			secretName := fmt.Sprintf("elastic-%s", f.Name)
			err = p.GetOrCreateSecret(secretName, constants.PlatformSystem, map[string][]byte{
				"ELASTIC_URL":      []byte(f.Elasticsearch.GetURL()),
				"ELASTIC_USERNAME": []byte(f.Elasticsearch.User),
				"ELASTIC_PASSWORD": []byte(password),
			})
			if err != nil {
				return errors.Wrap(err, "failed to create secret elastic")
			}
		}

		if f.Logstash != nil {
			secretName := fmt.Sprintf("logstash-%s", f.Name)
			var password string
			var err error
			_, password, err = p.GetEnvValue(f.Logstash.Password, "eck")
			if err != nil {
				_, password, err = p.GetEnvValue(f.Logstash.Password, metav1.NamespaceAll)
				if err != nil {
					return fmt.Errorf("unable to retrieve elasticsearch password for %s", f.Name)
				}
			}
			err = p.GetOrCreateSecret(secretName, constants.PlatformSystem, map[string][]byte{
				"LOGSTASH_URL":      []byte(f.Logstash.GetURL()),
				"LOGSTASH_USERNAME": []byte(f.Logstash.User),
				"LOGSTASH_PASSWORD": []byte(password),
			})
			if err != nil {
				return errors.Wrap(err, "Failed to create secret logstash")
			}
		}
	}

	return p.ApplySpecs(constants.PlatformSystem, "filebeat.yaml")
}
