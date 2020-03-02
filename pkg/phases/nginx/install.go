package nginx

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

func Install(platform *platform.Platform) error {
	if platform.Nginx != nil && platform.Nginx.Disabled {
		log.Debugf("Skipping nginx deployment")
		return nil
	}
	log.Infof("Installing Nginx Ingress Controller: %s", platform.Nginx.Version)
	if platform.Nginx == nil {
		platform.Nginx = &types.Nginx{}
	}
	if platform.Nginx.Version == "" {
		platform.Nginx.Version = "0.25.1.flanksource.1"
	}

	if platform.Nginx.RequestBodyBuffer == "" {
		platform.Nginx.RequestBodyBuffer = "16M"
	}

	if platform.Nginx.RequestBodyMax == "" {
		platform.Nginx.RequestBodyMax = "32M"
	}

	if err := platform.ApplySpecs("", "nginx.yml"); err != nil {
		log.Errorf("Error deploying nginx: %s\n", err)
	}

	if platform.OAuth2Proxy != nil && !platform.OAuth2Proxy.Disabled {
		return platform.ApplySpecs("", "nginx-oauth.yml")
	}
	return nil
}
