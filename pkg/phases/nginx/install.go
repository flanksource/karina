package nginx

import (
	"bytes"
	"fmt"
	"time"

	"github.com/flanksource/commons/utils"
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

	if platform.Nginx.RequestBodyBuffer == "" {
		platform.Nginx.RequestBodyBuffer = "16M"
	}

	if platform.Nginx.RequestBodyMax == "" {
		platform.Nginx.RequestBodyMax = "32M"
	}
	if platform.OAuth2Proxy != nil && !platform.OAuth2Proxy.Disabled {
		scripts, _ := platform.GetResourcesByDir("/nginx", "manifests")
		data := make(map[string]string)
		for name, file := range scripts {
			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(file); err != nil {
				return err
			}
			data[name] = buf.String()
		}
		if err := platform.CreateOrUpdateConfigMap("lua-scripts", Namespace, data); err != nil {
			return err
		}
		if !platform.HasSecret(Namespace, "oauth2-cookie-secret") {
			if err := platform.CreateOrUpdateSecret("oauth2-cookie-secret", Namespace, map[string][]byte{
				"secret": []byte(utils.RandomString(32)),
			}); err != nil {
				return err
			}
		}
		if err := platform.ApplySpecs(Namespace, "nginx-oauth.yaml"); err != nil {
			return err
		}
	} else {
		// need to create empty configmap for lua scripts
		if err := platform.CreateOrUpdateConfigMap("lua-scripts", Namespace, nil); err != nil {
			return err
		}
		if err := platform.DeleteSpecs(Namespace, "nginx-oauth.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	if err := platform.ApplySpecs(v1.NamespaceAll, "nginx.yaml"); err != nil {
		return err
	}
	// wait for the webhook to come up ready as otherwise subsequent ingress
	// creations will fail due to the validating webhook
	return platform.WaitForDeployment(Namespace, "nginx-ingress-webhook", 3*time.Minute)
}
