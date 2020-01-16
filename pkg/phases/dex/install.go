package dex

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace     = "dex"
	ConfigMapName = "dex"
	CertName      = "dex-cert"
	ConfigName    = "dex.cfg"
)

func Install(platform *platform.Platform) error {

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	if !platform.HasSecret(Namespace, CertName) {
		cert, err := platform.CreateIngressCertificate("dex")
		if err != nil {
			return err
		}
		if err := platform.CreateOrUpdateSecret(CertName, Namespace, cert.AsTLSSecret()); err != nil {
			return err
		}
	}

	cfg, _ := platform.Template("dex.cfg", "manifests")

	if err := platform.CreateOrUpdateConfigMap(ConfigMapName, Namespace, map[string]string{
		ConfigName: cfg,
	}); err != nil {
		return err
	}

	if err := platform.CreateOrUpdateSecret("ldap-account", Namespace, map[string][]byte{
		"AD_PASSWORD": []byte(platform.Ldap.Password),
		"AD_USERNAME": []byte(platform.Ldap.Username),
	}); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "dex.yaml")
}
