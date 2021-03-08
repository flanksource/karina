package certmanager

import (
	"context"
	"fmt"
	"strings"

	"encoding/json"
	"github.com/blang/semver/v4"
	"github.com/flanksource/commons/certs"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	acmev1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	ccmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace                 = "cert-manager"
	IngressCA                 = "ingress-ca"
	DefaultIssuerCA           = "default-issuer-ca"
	VaultTokenName            = "vault-token"
	Route53Name               = "route53-credentials"
	LetsencryptPrivateKeyName = "letsencrypt-issuer-account-key"
	SecretKeyName             = "AWS_SECRET_ACCESS_KEY"
	WebhookService            = "cert-manager-webhook"
)

func PreInstall(p *platform.Platform) error {
	client, err := p.Client.GetClientset()
	if err != nil {
		return err
	}

	// Cert manager is a core component and multiple other components depend on it so it cannot be disabled
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	// First generate or load he CA bundle for the default-issuer which is a dependency for
	// creating CRD's with webhooks.

	caBundle, _ := p.GetSecretValue(Namespace, DefaultIssuerCA, "tls.crt")

	if caBundle == nil {
		ca := p.NewSelfSigned("default-issuer")
		if err := p.Apply(Namespace, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: Namespace,
				Name:      DefaultIssuerCA,
				Annotations: map[string]string{
					certmanager.AllowsInjectionFromSecretAnnotation: "true",
				},
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			Data: ca.AsTLSSecret(),
		}); err != nil {
			return err
		}
		p.CertManager.DefaultIssuerCA = string(ca.AsTLSSecret()["tls.crt"])
	} else {
		p.CertManager.DefaultIssuerCA = string(caBundle)
	}

	deployments := client.AppsV1().Deployments(Namespace)
	deployment, err := deployments.Get(context.TODO(), "cert-manager", metav1.GetOptions{})
	if err != nil {
		// cert-manager is not installed, nothing todo
		return nil
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		currentVer := strings.TrimLeft(strings.Split(container.Image, ":")[1], "v")
		v, _ := semver.Parse(currentVer)
		preGA, _ := semver.ParseRange("<1.0.0")
		if preGA(v) {
			p.Debugf("Upgrading cert-manager from %s -> %s, deleting existing deployment ", currentVer, p.PlatformConfig.CertManager.Version)
			break
		} else {
			return nil
		}
	}

	if err := p.DeleteByKind(constants.ValidatingWebhookConfiguration, v1.NamespaceAll, WebhookService); err != nil {
		return err
	}
	for _, name := range []string{"cert-manager", "cert-manager-webhook", "cert-manager-cainjector"} {
		if err := deployments.Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func Install(p *platform.Platform) error {
	//remove old mutating webhooks
	_ = p.DeleteByKind(constants.MutatingWebhookConfiguration, v1.NamespaceAll, WebhookService)

	if err := p.ApplySpecs("", "cert-manager-deploy.yaml", "cert-manager-monitor.yaml.raw"); err != nil {
		return fmt.Errorf("failed to deploy cert-manager: %v", err)
	}

	if p.DryRun {
		return nil
	}

	if err := createIngressCA(p); err != nil {
		return err
	}

	ca, err := p.CreateOrGetWebhookCertificate(Namespace, WebhookService)
	if err != nil {
		return err
	}

	if err := p.ApplySpecs("", "cert-manager-webhook.yaml"); err != nil {
		return fmt.Errorf("failed to deploy cert-manager webhook: %v", err)
	}

	webhooks, err := p.CreateWebhookBuilder(Namespace, WebhookService, ca)
	if err != nil {
		return err
	}

	webhooks = webhooks.NewHook("webhook.cert-manager.io", "/validate").
		WithoutNamespaceLabel("cert-manager.io/disable-validation", "true").
		Match([]string{"cert-manager.io", "acme.cert-manager.io"}, []string{"*"}, []string{"*/*"}).
		Add()

	return p.Apply(Namespace, webhooks.Build())
}

func createIngressCA(p *platform.Platform) error {
	if issuer, _ := p.GetByKind(certmanager.ClusterIssuerKind, v1.NamespaceAll, IngressCA); issuer != nil {
		// We only deploy the ingress-ca once, and then forget about it, this is for 2 reasons:
		// 1) Not polluting the audit log with unnecessary read requests to the CA Key
		// 2) Allow running deployment with less secrets once the CA is deployed
		p.Tracef("Ingress CA already configured, skipping")
		return nil
	}
	var issuerConfig certmanager.IssuerConfig

	if p.CertManager.Vault != nil {
		// TODO(moshloop): delete previously imported CA
		p.Infof("Configuring Cert Manager ClusterIssuer to use Vault: ingress-ca")
		if err := p.CreateOrUpdateSecret(VaultTokenName, Namespace, map[string][]byte{
			"token": []byte(p.CertManager.Vault.Token),
		}); err != nil {
			return err
		}
		issuerConfig = certmanager.IssuerConfig{
			CA: nil,
			Vault: &certmanager.VaultIssuer{
				Server:   p.CertManager.Vault.Address,
				CABundle: p.GetIngressCA().GetPublicChain()[0].EncodedCertificate(),
				Path:     p.CertManager.Vault.Path,
				Auth: certmanager.VaultAuth{
					TokenSecretRef: &ccmetav1.SecretKeySelector{
						Key: "token",
						LocalObjectReference: ccmetav1.LocalObjectReference{
							Name: VaultTokenName,
						},
					},
				},
			},
		}
	} else if p.CertManager.Letsencrypt != nil {
		p.Infof("Configuring Cert Manager ClusterIssuer to use Letsencrypt: ingress-ca")
		if p.DNS.SecretKey != "" {
			if err := p.CreateOrUpdateSecret(Route53Name, Namespace, map[string][]byte{
				SecretKeyName: []byte(p.DNS.SecretKey),
			}); err != nil {
				return err
			}
		}
		var solver acmev1.ACMEChallengeSolver
		if p.DNS.Type == "route53" {
			solver = acmev1.ACMEChallengeSolver{
				DNS01: &acmev1.ACMEChallengeSolverDNS01{
					Route53: &acmev1.ACMEIssuerDNS01ProviderRoute53{
						Region:       p.DNS.Region,
						HostedZoneID: p.DNS.Zone,
						AccessKeyID:  p.DNS.AccessKey,
						SecretAccessKey: ccmetav1.SecretKeySelector{
							LocalObjectReference: ccmetav1.LocalObjectReference{
								Name: Route53Name,
							},
						},
					},
				},
			}
		} else {
			solver = acmev1.ACMEChallengeSolver{
				HTTP01: &acmev1.ACMEChallengeSolverHTTP01{
					Ingress: &acmev1.ACMEChallengeSolverHTTP01Ingress{
						Name: "ingress",
					},
					// Type:
				},
			}
		}
		var server string
		if p.CertManager.Letsencrypt.URL == "" {
			server = "https://acme-v02.api.letsencrypt.org/directory"
		} else {
			server = p.CertManager.Letsencrypt.URL
		}
		issuerConfig = certmanager.IssuerConfig{
			ACME: &acmev1.ACMEIssuer{
				Server:  server,
				Email:   p.CertManager.Letsencrypt.Email,
				Solvers: []acmev1.ACMEChallengeSolver{solver},
			},
		}
	} else {
		p.Infof("Importing Ingress CA as a Cert Manager ClusterIssuer: ingress-ca")
		ingress := p.GetIngressCA()
		switch ingress := ingress.(type) {
		case *certs.Certificate:
			if err := p.CreateOrUpdateSecret(IngressCA, Namespace, ingress.AsTLSSecret()); err != nil {
				return err
			}
			issuerConfig = certmanager.IssuerConfig{
				Vault: nil,
				CA: &certmanager.CAIssuer{
					SecretName: IngressCA,
				},
			}
			j, _ := json.Marshal(issuerConfig)
			fmt.Println(j)
		default:
			return fmt.Errorf("unknown cert type:%v", ingress)
		}
	}

	return p.Apply(Namespace, &certmanager.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssuer",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      IngressCA,
			Namespace: Namespace,
		},
		Spec: certmanager.IssuerSpec{
			IssuerConfig: issuerConfig,
		},
	})
}
