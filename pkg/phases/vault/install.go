package vault

import (
	"strconv"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "vault"
)

func Deploy(p *platform.Platform) error {
	if p.Vault == nil || p.Vault.Disabled {
		if err := p.DeleteSpecs(Namespace, "vault.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
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

	if p.Vault.Consul.Bucket != "" {
		if err := p.CreateOrUpdateSecret("consul-backup-config", Namespace, map[string][]byte{
			"AWS_REGION":              []byte(p.S3.Region),
			"AWS_ACCESS_KEY_ID":       []byte(p.S3.AccessKey),
			"AWS_SECRET_ACCESS_KEY":   []byte(p.S3.SecretKey),
			"AWS_ENDPOINT":            []byte(p.S3.Endpoint),
			"AWS_S3_FORCE_PATH_STYLE": []byte(strconv.FormatBool(p.S3.UsePathStyle)),
		}); err != nil {
			return err
		}
		if err := p.GetOrCreateBucket(p.Vault.Consul.Bucket); err != nil {
			return err
		}
	}

	return p.ApplySpecs(Namespace, "vault.yaml")
}
