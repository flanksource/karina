package opa

import (
	"crypto/x509"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
)

const (
	Namespace    = "opa"
	SecretNameCA = "opa-ca"
)

func Install(platform *platform.Platform) error {
	if platform.OPA != nil && platform.OPA.Disabled {
		return nil
	} else if platform.OPA == nil {
		platform.OPA = &types.OPA{
			Version:         "0.13.5",
			KubeMgmtVersion: "0.8"}
	}

	platform.GetKubectl()("create ns opa")
	platform.GetKubectl()("label ns opa app=opa")

	log.Infof("Creating a CA for %s\n", "opa.")
	opaCA, err := utils.NewCertificateAuthority("opa")

	platform.CreateOrUpdateSecret(SecretNameCA, Namespace, map[string][]byte{
		"ca.crt": opaCA.EncodedCertificate(),
		"ca.key": opaCA.EncodedPrivateKey(),
	})

	log.Infof("Updating secret %s\n", "opa.")
	annotations := map[string]string{"cert-manager.io/allow-direct-injection": "true"}
	platform.Annotate("secrets", SecretNameCA, Namespace, annotations)

	log.Infof("Creating tls creds for %s\n", "opa.")

	usages := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	cfg := utils.Config{
		CommonName: "opa.opa.svc",
		Usages:     usages,
	}

	opaPrivateKey, _ := utils.NewPrivateKey()

	opaCert, err := cfg.NewSignedCert(opaPrivateKey, opaCA.X509, opaCA.PrivateKey)
	if err != nil {
		return err
	}

	opaCertificate := utils.Certificate{X509: opaCert, PrivateKey: opaPrivateKey}

	platform.CreateOrUpdateSecret("opa-server", Namespace, map[string][]byte{
		"tls.crt": opaCertificate.EncodedCertificate(),
		"tls.key": opaCertificate.EncodedPrivateKey(),
	})
	return platform.ApplySpecs(Namespace, "opa.yaml")
}
