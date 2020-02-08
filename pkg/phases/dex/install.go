package dex

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

const (
	Namespace     = "dex"
	ConfigMapName = "dex"
	CertName      = "dex-cert"
	ConfigName    = "dex.cfg"
)

func Install(platform *platform.Platform) error {

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		log.Tracef("Install: Failed to create/update namespace: %s", err)
		return err
	}

	if !platform.HasSecret(Namespace, CertName) {
		cert, err := platform.CreateIngressCertificate("dex")
		if err != nil {
			log.Tracef("Install: Failed to create ingress certificate: %s", err)
			return err
		}
		if err := platform.CreateOrUpdateSecret(CertName, Namespace, cert.AsTLSSecret()); err != nil {
			log.Tracef("Install: Failed to create/update secret: %s", err)
			return err
		}
	}

	cfg, _ := platform.Template("dex.cfg", "manifests")

	if err := platform.CreateOrUpdateConfigMap(ConfigMapName, Namespace, map[string]string{
		ConfigName: cfg,
	}); err != nil {
		log.Tracef("Install: Failed to create/update configmap: %s", err)
		return err
	}

	if err := platform.CreateOrUpdateSecret("ldap-account", Namespace, map[string][]byte{
		"AD_PASSWORD": []byte(platform.Ldap.Password),
		"AD_USERNAME": []byte(platform.Ldap.Username),
	}); err != nil {
		log.Tracef("Install: Failed to create/update secret: %s", err)
		return err
	}

	return platform.ApplySpecs(Namespace, "dex.yaml")
}
