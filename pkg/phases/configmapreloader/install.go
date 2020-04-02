package configmapreloader

import (
	"github.com/flanksource/commons/utils"
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "configmap-reloader"
)

func Deploy(p *platform.Platform) error {
	if p.ConfigMapReloader.Disabled {
		log.Infof("Skipping deployment of configmap-reloader, it is disabled")
		return nil
	}

	if p.ConfigMapReloader.Version == "" {
		p.ConfigMapReloader.Version = "v0.0.56"
	} else {
		p.ConfigMapReloader.Version = utils.NormalizeVersion(p.ConfigMapReloader.Version)
	}

	log.Infof("Deploying configmap-reloader %s", p.ConfigMapReloader.Version)

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return p.ApplySpecs("", "configmap-reloader.yaml")
}
