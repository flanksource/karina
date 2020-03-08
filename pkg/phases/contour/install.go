package contour

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func Deploy(p *platform.Platform) error {
	if p.Contour != nil || !p.Contour.Disabled {
		log.Infof("Deploying Contour %s", p.Contour.Version)
	} else {
		log.Infof("Skipping deployment of Contour, Nginx ingress controller will be used")
		return nil
	}
	if err := p.ApplySpecs("", "contour.yaml"); err != nil {
		log.Warnf("Failed to deploy contour %v", err)
	}
	return nil
}
