package vault

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "vault"
)

func Deploy(p *platform.Platform) error {
	if p.Vault == nil || p.Vault.Disabled {
		log.Infof("Skipping deployment of vault, it is disabled")
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	if !p.HasSecret(Namespace, "vault-tls") {
		//FIXME(moshloop) does this need an internal or external cert?
		crt, err := p.CreateInternalCertificate("vault", Namespace, "cluster.local")
		if err != nil {
			return err
		}
		if err := p.CreateOrUpdateSecret("vault-tls", Namespace, crt.AsTLSSecret()); err != nil {
			return err
		}
	}

	if err := p.CreateOrUpdateSecret("kms", Namespace, map[string][]byte{
		"AWS_REGION":               []byte(p.Vault.Region),
		"AWS_ACCESS_KEY_ID":        []byte(p.Vault.AccessKey),
		"AWS_SECRET_ACCESS_KEY":    []byte(p.Vault.SecretKey),
		"VAULT_AWSKMS_SEAL_KEY_ID": []byte(p.Vault.KmsKeyID),
	}); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "vault.yaml")
}
