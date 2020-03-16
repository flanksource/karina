package certmanager

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/commons/certs"
	"github.com/moshloop/platform-cli/pkg/api/certmanager"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "cert-manager"
	IngressCA = "ingress-ca"
	VaultTokenName = "vault-token"
)

func Install(platform *platform.Platform) error {
	if platform.CertManager != nil && platform.CertManager.Disabled {
		log.Infof("Cert Manager is disabled, skipping")
		return nil
	}
	log.Infof("Installing CertMananager")
	if err := platform.ApplySpecs("", "cert-manager-crd.yaml"); err != nil {
		log.Errorf("Error deploying cert manager CRDs: %s\n", err)
	}

	// the cert-manager webhook can take time to deploy, so we deploy it once ignoring any errors
	// wait for 180s for the namespace to be ready, deploy again (usually a no-op) and only then report errors
	var _ = platform.ApplySpecs("", "cert-manager-deploy.yaml")
	platform.GetKubectl()("wait --timeout=300s --for=condition=Available apiservice v1beta1.webhook.cert-manager.io")
	platform.WaitForNamespace(Namespace, 180*time.Second)
	var issuerConfig certmanager.IssuerConfig
	if platform.Vault == nil || platform.Vault.Address == "" {
		log.Infof("Importing Ingress CA as a Cert Manager ClusterIssuer: ingress-ca")
		ingress := platform.GetIngressCA()
		switch ingress := ingress.(type) {
		case *certs.Certificate:
			if err := platform.CreateOrUpdateSecret(IngressCA, Namespace, ingress.AsTLSSecret()); err != nil {
				return err
			}
			issuerConfig = certmanager.IssuerConfig{
				CA: &certmanager.CAIssuer{
					SecretName: IngressCA,
				},
			}
		default:
			return fmt.Errorf("Unknown cert type:%v", ingress)
		}
	} else {
		if err := platform.CreateOrUpdateSecret(VaultTokenName, Namespace, map[string][]byte{
			"token": []byte(platform.Vault.Token),
		}); err != nil {
			return err
		}
		issuerConfig = certmanager.IssuerConfig{
			Vault: &certmanager.VaultIssuer{
				Server: platform.Vault.Address,
				Auth: certmanager.VaultAuth{
					TokenSecretRef: &certmanager.SecretKeySelector{
						Key: "token",
						LocalObjectReference: certmanager.LocalObjectReference{
							Name: VaultTokenName,
						},
					},
				},
			},
		}
	}

	if err := platform.Apply(Namespace, &certmanager.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssuer",
			APIVersion: "cert-manager.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      IngressCA,
			Namespace: Namespace,
		},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: issuerConfig,
		},
	}); err != nil {
		return err
	}

	return nil

}
