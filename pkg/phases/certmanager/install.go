package certmanager

import (
	"fmt"
	"time"

	"github.com/flanksource/commons/certs"
	"github.com/moshloop/platform-cli/pkg/api/certmanager"
	"github.com/moshloop/platform-cli/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace      = "cert-manager"
	IngressCA      = "ingress-ca"
	VaultTokenName = "vault-token"
)

func Install(platform *platform.Platform) error {
	// Cert manager is a core component and multiple other components depend on it
	// so it cannot be disabled
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	platform.Infof("Installing CertMananager")
	if err := platform.ApplySpecs("", "cert-manager-crd.yaml"); err != nil {
		return err
	}

	if err := platform.ApplySpecs("", "cert-manager-deploy.yaml"); err != nil {
		return err
	}

	platform.WaitForNamespace(Namespace, 180*time.Second)

	var issuerConfig certmanager.IssuerConfig
	if platform.CertManager.Vault == nil {
		platform.Infof("Importing Ingress CA as a Cert Manager ClusterIssuer: ingress-ca")
		ingress := platform.GetIngressCA()
		switch ingress := ingress.(type) {
		case *certs.Certificate:
			if err := platform.CreateOrUpdateSecret(IngressCA, Namespace, ingress.AsTLSSecret()); err != nil {
				return err
			}
			issuerConfig = certmanager.IssuerConfig{
				Vault: nil,
				CA: &certmanager.CAIssuer{
					SecretName: IngressCA,
				},
			}
		default:
			return fmt.Errorf("unknown cert type:%v", ingress)
		}
	} else {
		// TODO(moshloop): delete previously imported CA

		platform.Infof("Configuring Cert Manager ClusterIssuer to use Vault: ingress-ca")
		if err := platform.CreateOrUpdateSecret(VaultTokenName, Namespace, map[string][]byte{
			"token": []byte(platform.CertManager.Vault.Token),
		}); err != nil {
			return err
		}
		issuerConfig = certmanager.IssuerConfig{
			CA: nil,
			Vault: &certmanager.VaultIssuer{
				Server:   platform.CertManager.Vault.Address,
				CABundle: platform.GetIngressCA().GetPublicChain()[0].EncodedCertificate(),
				Path:     platform.CertManager.Vault.Path,
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
