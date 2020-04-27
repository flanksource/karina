package dex

import (
	"fmt"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace     = "dex"
	ConfigMapName = "dex"
	CertName      = "dex-cert"
	ConfigName    = "dex.cfg"
)

func dexLabels() map[string]string {
	return map[string]string{
		"app": "dex",
	}
}

func Install(platform *platform.Platform) error {
	if err := platform.CreateOrUpdateNamespace(Namespace, dexLabels(), nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := platform.CreateTLSSecret(Namespace, "dex", CertName); err != nil {
		return err
	}

	cfg, _ := platform.Template("dex.cfg", "manifests")

	if err := platform.CreateOrUpdateConfigMap(ConfigMapName, Namespace, map[string]string{
		ConfigName: cfg,
	}); err != nil {
		return fmt.Errorf("install: failed to create/update configmap: %v", err)
	}

	if err := platform.CreateOrUpdateSecret("ldap-account", Namespace, map[string][]byte{
		"AD_PASSWORD": []byte(platform.Ldap.Password),
		"AD_USERNAME": []byte(platform.Ldap.Username),
	}); err != nil {
		return fmt.Errorf("install: failed to create/update secret: %v", err)
	}

	return platform.ApplySpecs(Namespace, "dex.yaml")
}
