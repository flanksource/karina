package nginx

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	v1 "k8s.io/api/core/v1"
)

const (
	Namespace = "ingress-nginx"
)

func Install(platform *platform.Platform) error {
	if platform.Nginx != nil && platform.Nginx.Disabled {
		if err := platform.DeleteSpecs(v1.NamespaceAll, "nginx.yaml", "nginx-oauth.yaml"); err != nil {
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

	if platform.Nginx.RequestBodyBuffer == "" {
		platform.Nginx.RequestBodyBuffer = "16M"
	}

	if platform.Nginx.RequestBodyMax == "" {
		platform.Nginx.RequestBodyMax = "32M"
	}

	if err := platform.ApplySpecs(v1.NamespaceAll, "nginx.yaml"); err != nil {
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
