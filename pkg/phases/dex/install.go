package dex

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace     = "dex"
	ConfigMapName = "dex"
	ConfigName    = "dex.cfg"
)

func Install(platform *platform.Platform) error {

	platform.GetKubectl()("create ns dex")

	openid := platform.Certificates.OpenID.ToCert()
	log.Infof("Creating dex cert for %s\n", "dex."+platform.Domain)
	cert, err := openid.CreateCertificate("dex."+platform.Domain, "")
	if err != nil {
		return err
	}

	if err := platform.CreateOrUpdateSecret("dex-cert", Namespace, map[string][]byte{
		"tls.crt": cert.EncodedCertificate(),
		"tls.key": cert.EncodedPrivateKey(),
	}); err != nil {
		return err
	}

	cfg, _ := platform.Template("dex.cfg")

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
