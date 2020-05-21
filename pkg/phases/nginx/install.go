package nginx

import (
	"fmt"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

const (
	Namespace = "ingress-nginx"
)

func Install(platform *platform.Platform) error {
	if platform.Nginx != nil && platform.Nginx.Disabled {
		if err := platform.DeleteSpecs(Namespace, "nginx.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if platform.Nginx == nil {
		platform.Nginx = &types.Nginx{}
	}
	if platform.Nginx.Version == "" {
		platform.Nginx.Version = "0.25.1.flanksource.1"
	}
	platform.Infof("Installing Nginx Ingress Controller: %s", platform.Nginx.Version)

	if platform.Nginx.RequestBodyBuffer == "" {
		platform.Nginx.RequestBodyBuffer = "16M"
	}

	if platform.Nginx.RequestBodyMax == "" {
		platform.Nginx.RequestBodyMax = "32M"
	}

	if err := platform.ApplySpecs(Namespace, "nginx.yaml"); err != nil {
		platform.Errorf("Error deploying nginx: %s\n", err)
	}

	if platform.OAuth2Proxy != nil && !platform.OAuth2Proxy.Disabled {
		return platform.ApplySpecs(Namespace, "nginx-oauth.yaml")
	}
	if err := platform.DeleteSpecs(Namespace, "nginx-oauth.yaml"); err != nil {
		platform.Warnf("failed to delete specs: %v", err)
	}
	return nil
}
