package stubs

import (
	"github.com/flanksource/karina/pkg/platform"
)

const MinioNamespace = "minio"
const LdapNamespace = "ldap"

func Install(platform *platform.Platform) error {
	if platform.S3.E2E.Minio {
		if err := platform.CreateOrUpdateNamespace(MinioNamespace, nil, nil); err != nil {
			return err
		}
		if err := platform.ApplySpecs(MinioNamespace, "minio.yaml"); err != nil {
			platform.Errorf("Error deploying minio: %s\n", err)
		}
	} else {
		if err := platform.DeleteSpecs(MinioNamespace, "minio.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	if platform.Ldap != nil && !platform.Ldap.Disabled && platform.Ldap.E2E.Mock {
		if err := platform.CreateOrUpdateNamespace(LdapNamespace, nil, nil); err != nil {
			return err
		}
		if err := platform.ApplySpecs(LdapNamespace, "apacheds.yaml"); err != nil {
			platform.Errorf("Error deploying apacheds: %s\n", err)
		}
	} else {
		if err := platform.DeleteSpecs(LdapNamespace, "apacheds.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}
	return nil
}
