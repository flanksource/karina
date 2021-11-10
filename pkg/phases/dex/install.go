package dex

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace     = "dex"
	ConfigMapName = "dex"
	CertName      = "dex-cert"
	ConfigName    = "dex.cfg"
)

func Install(platform *platform.Platform) error {
	if platform.Dex.Version == "" {
		platform.Dex.Version = "v2.27.0"
	}
	if platform.Dex.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "dex.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := platform.CreateTLSSecret(Namespace, "dex", CertName); err != nil {
		return err
	}

	cfg, err := platform.Template("dex.cfg", "manifests")
	if err != nil {
		return fmt.Errorf("install: failed to template configmap: %v", err)
	}

	if err := platform.CreateOrUpdateConfigMap(ConfigMapName, Namespace, map[string]string{
		ConfigName: cfg,
	}); err != nil {
		return fmt.Errorf("install: failed to create/update configmap: %v", err)
	}

	if platform.Dex.Google.ClientID != "" {
		if err := platform.CreateOrUpdateSecret("google-account", Namespace, map[string][]byte{
			"GOOGLE_CLIENT_ID":     []byte(platform.Dex.Google.ClientID),
			"GOOGLE_CLIENT_SECRET": []byte(platform.Dex.Google.ClientSecret),
		}); err != nil {
			return fmt.Errorf("install: failed to create/update google secret: %v", err)
		}
	}
	if platform.Dex.Github.ClientID != "" {
		if err := platform.CreateOrUpdateSecret("github-account", Namespace, map[string][]byte{
			"GITHUB_CLIENT_ID":     []byte(platform.Dex.Github.ClientID),
			"GITHUB_CLIENT_SECRET": []byte(platform.Dex.Github.ClientSecret),
		}); err != nil {
			return fmt.Errorf("install: failed to create/update github secret: %v", err)
		}
	}
	if platform.Dex.Gitlab.ApplicationID != "" {
		if err := platform.CreateOrUpdateSecret("gitlab-account", Namespace, map[string][]byte{
			"GITLAB_APPLICATION_ID": []byte(platform.Dex.Gitlab.ApplicationID),
			"GITLAB_CLIENT_SECRET":  []byte(platform.Dex.Gitlab.ClientSecret),
		}); err != nil {
			return fmt.Errorf("install: failed to create/update gitlab secret: %v", err)
		}
	}
	if !platform.Ldap.Disabled {
		if err := platform.CreateOrUpdateSecret("ldap-account", Namespace, map[string][]byte{
			"AD_PASSWORD": []byte(platform.Ldap.Password),
			"AD_USERNAME": []byte(platform.Ldap.Username),
		}); err != nil {
			return fmt.Errorf("install: failed to create/update ldap secret: %v", err)
		}
	}

	return platform.ApplySpecs(Namespace, "dex.yaml")
}
