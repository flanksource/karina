package certmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/flanksource/commons/certs"
	"github.com/flanksource/karina/pkg/api/certmanager"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace      = "cert-manager"
	IngressCA      = "ingress-ca"
	VaultTokenName = "vault-token"
	Route53Name    = "route53-credentials"
	SecretKeyName  = "AWS_SECRET_ACCESS_KEY"
	WebhookService = "cert-manager-webhook"
)

func PreInstall(platform *platform.Platform) error {
	client, err := platform.Client.GetClientset()
	if err != nil {
		return err
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
			platform.Debugf("Upgrading cert-manager from %s -> %s, deleting existing deployment ", currentVer, platform.PlatformConfig.CertManager.Version)
			break
		} else {
			return nil
		}
	}

	if err := platform.DeleteByKind(constants.ValidatingWebhookConfiguration, v1.NamespaceAll, WebhookService); err != nil {
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
	// Cert manager is a core component and multiple other components depend on it
	// so it cannot be disabled
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	if err := p.ApplySpecs("", "cert-manager-deploy.yaml"); err != nil {
		return fmt.Errorf("failed to deploy cert-manager: %v", err)
	}
	if err := p.ApplySpecs("", "cert-manager-monitor.yaml.raw"); err != nil {
		return fmt.Errorf("failed to deploy cert-manager alerts: %v", err)
	}

	if p.DryRun {
		return nil
	}

	if err := createDefaultIssuer(p); err != nil {
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

func createDefaultIssuer(p *platform.Platform) error {
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
					TokenSecretRef: &certmanager.SecretKeySelector{
						Key: "token",
						LocalObjectReference: certmanager.LocalObjectReference{
							Name: VaultTokenName,
						},
					},
				},
			},
		}
	} else if p.CertManager.Letsencrypt != nil {
		if p.DNS.SecretKey != "" {
			if err := p.CreateOrUpdateSecret(Route53Name, Namespace, map[string][]byte{
				SecretKeyName: []byte(p.DNS.SecretKey),
			}); err != nil {
				return err
			}
		}
		var solver certmanager.Solver
		if p.DNS.Type == "route53" {
			solver = certmanager.Solver{
				DNS01: certmanager.DNS01{
					Route53: certmanager.Route53{
						Region:       p.DNS.Region,
						HostedZoneID: p.DNS.Zone,
						AccessKeyID:  p.DNS.AccessKey,
						SecretAccessKeyRef: certmanager.SecretKeySelector{
							LocalObjectReference: certmanager.LocalObjectReference{
								Name: Route53Name,
							},
							Key: SecretKeyName,
						},
					},
				},
			}
		} else {
			solver = certmanager.Solver{
				HTTP01: certmanager.HTTP01{
					Type: "ingress",
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
			Letsencrypt: &certmanager.LetsencryptIssuer{
				Server:  server,
				Email:   p.CertManager.Letsencrypt.Email,
				Solvers: []certmanager.Solver{solver},
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
