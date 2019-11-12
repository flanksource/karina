package opa

import (
	"crypto/x509"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	k8s, err := platform.GetClientset()
	if err != nil {
		return err
	}
	ns := k8s.CoreV1().Namespaces()
	secrets := k8s.CoreV1().Secrets(Namespace)

	if _, err := ns.Get(Namespace, metav1.GetOptions{}); errors.IsNotFound(err) {
		opa := v1.Namespace{}
		opa.Name = "opa"
		opa.Labels = map[string]string{"app": "opa"}
		if _, err := ns.Create(&opa); err != nil {
			return err
		}
	}

	if _, err := secrets.Get(SecretNameCA, metav1.GetOptions{}); errors.IsNotFound(err) {

		log.Infof("Creating a CA for %s\n", "opa.")
		opaCA, err := utils.NewCertificateAuthority("opa")

		if err := platform.CreateOrUpdateSecret(SecretNameCA, Namespace, map[string][]byte{
			"ca.crt": opaCA.EncodedCertificate(),
			"ca.key": opaCA.EncodedPrivateKey(),
		}); err != nil {
			return err
		}

		if err := platform.Annotate("secrets",
			SecretNameCA,
			Namespace,
			map[string]string{"cert-manager.io/allow-direct-injection": "true"}); err != nil {
			return err
		}

		log.Debugf("Creating tls creds for opa")

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

		if err := platform.CreateOrUpdateSecret("opa-server", Namespace, map[string][]byte{
			"tls.crt": opaCertificate.EncodedCertificate(),
			"tls.key": opaCertificate.EncodedPrivateKey(),
		}); err != nil {
			return err
		}
	}
	return platform.ApplySpecs(Namespace, "opa.yaml")
}
