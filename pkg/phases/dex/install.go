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
	openid := platform.Certificates.OpenID.ToCert()
	log.Infof("Creating dex cert for %s\n", "dex."+platform.Domain)
	cert, err := openid.CreateCertificate("dex."+platform.Domain, "")
	if err != nil {
		return err
	}
	platform.CreateOrUpdateSecret("dex-cert", Namespace, map[string][]byte{
		"tls.crt": cert.EncodedCertificate(),
		"tls.key": cert.EncodedPrivateKey(),
	})

	cfg, _ := platform.Template("dex.cfg")

	platform.CreateOrUpdateConfigMap(ConfigMapName, Namespace, map[string]string{
		ConfigName: cfg,
	})

	platform.CreateOrUpdateSecret("ldap-account", Namespace, map[string][]byte{
		"AD_PASSWORD": []byte(platform.Ldap.Password),
		"AD_USERNAME": []byte(platform.Ldap.Username),
	})

	return platform.ApplySpecs(Namespace, "dex.yaml")
}
