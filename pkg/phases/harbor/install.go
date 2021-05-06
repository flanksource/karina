package harbor

import (
	"fmt"

	pgapi "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/konfigadm/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	HaborRegistryUsername = "harbor_registry_user"
)

var manifests = []string{"core", "portal", "registry", "exporter", "redis", "jobservice", "chartmuseum", "trivy"}

func Deploy(p *platform.Platform) error {
	if p.Harbor == nil || p.Harbor.Disabled {
		for _, spec := range manifests {
			if err := p.DeleteSpecs("", fmt.Sprintf("harbor/%s.yaml", spec)); err != nil {
				p.Warnf("failed to delete specs: %v", err)
			}
		}

		return nil
	}

	if p.Harbor.S3 == nil {
		p.Harbor.S3 = &p.S3.S3Connection
	}

	if p.Harbor.Bucket != "" {
		if err := p.GetOrCreateBucketFor(*p.Harbor.S3, p.Harbor.Bucket); err != nil {
			return err
		}
	}
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}
	var nonce string
	if p.HasSecret(Namespace, "harbor-secret") {
		nonce = string((*p.GetSecret(Namespace, "harbor-secret"))["secret"])
	} else {
		nonce = utils.RandomString(16)
		if err := p.CreateOrUpdateSecret("harbor-secret", Namespace, map[string][]byte{
			"secret": []byte(nonce),
		}); err != nil {
			return err
		}
	}
	harborCore := p.GetSecret(Namespace, "harbor-core")
	harborRegistry := p.GetSecret(Namespace, "harbor-registry")
	var registryPassword []byte
	var registryHtPassword []byte
	var err error
	if harborCore != nil && (*harborCore)["REGISTRY_CREDENTIAL_PASSWORD"] != nil {
		registryPassword = (*harborCore)["REGISTRY_CREDENTIAL_PASSWORD"]
		registryHtPassword = (*harborRegistry)["REGISTRY_HTPASSWD"]
	} else {
		registryPassword = []byte(utils.RandomString(16))
		registryHtPassword, err = getHtPasswd(string(registryPassword))
		if err != nil {
			return err
		}
	}
	var csrfKey []byte

	if harborCore != nil && (*harborCore)["CSRF_KEY"] != nil {
		csrfKey = (*harborCore)["CSRF_KEY"]
	} else {
		csrfKey = []byte(utils.RandomString(32))
	}

	if !p.HasSecret(Namespace, "token-key") {
		var tokenKey, tokenCrt []byte
		if harborCore != nil && (*harborCore)["tls.key"] != nil {
			// migrate key over from existing installation
			tokenKey = (*harborCore)["tls.key"]
			tokenCrt = (*harborCore)["tls.crt"]
		} else {
			token := p.NewSelfSigned("registry-token")
			tokenKey = token.EncodedPrivateKey()
			tokenCrt = token.EncodedCertificate()
		}
		if err := p.CreateOrUpdateSecret("token-key", Namespace, map[string][]byte{
			"tls.key": tokenKey,
			"tls.crt": tokenCrt,
		}); err != nil {
			return err
		}
	}

	if p.Harbor.DB == nil {
		if !p.ApplyDryRun {
			dbConfig := pgapi.NewClusterConfig(dbCluster, dbNames...)
			db, err := postgresoperator.GetOrCreateDB(p, dbConfig)
			if err != nil {
				return fmt.Errorf("deploy: failed to get/update db: %v", err)
			}
			p.Harbor.DB = db
		} else {
			p.Debugf("Creating postgres database %s", dbCluster)
		}
	}

	if !p.ApplyDryRun {
		chartSecret := map[string][]byte{
			"CACHE_REDIS_PASSWORD": []byte{},
		}
		if p.Harbor.ChartPVC == "" {
			chartSecret["AWS_SECRET_ACCESS_KEY"] = []byte(p.Harbor.S3.SecretKey)
		}
		if err := p.CreateOrUpdateSecret("harbor-chartmuseum", Namespace, chartSecret); err != nil {
			return err
		}

		if err := p.CreateOrUpdateSecret("harbor-trivy", Namespace, map[string][]byte{
			"gitHubToken": []byte{},
			"redisURL":    []byte("redis://harbor-redis:6379/4"),
		}); err != nil {
			return err
		}

		if err := p.CreateOrUpdateSecret("harbor-core", Namespace, map[string][]byte{
			"HARBOR_ADMIN_PASSWORD":        []byte(p.Harbor.AdminPassword),
			"POSTGRESQL_PASSWORD":          []byte(p.Harbor.DB.Password),
			"REGISTRY_CREDENTIAL_PASSWORD": registryPassword,
			"secretKey":                    []byte("not-a-secure-key"),
			"secret":                       []byte(nonce),
			"CSRF_KEY":                     csrfKey,
		}); err != nil {
			return err
		}

		registrySecret := map[string][]byte{
			"REGISTRY_HTPASSWD":       registryHtPassword,
			"REGISTRY_HTTP_SECRET":    []byte(nonce),
			"REGISTRY_REDIS_PASSWORD": []byte(""),
		}
		if p.Harbor.RegistryPVC == "" {
			registrySecret["REGISTRY_STORAGE_S3_ACCESSKEY"] = []byte(p.Harbor.S3.AccessKey)
			registrySecret["REGISTRY_STORAGE_S3_SECRETKEY"] = []byte(p.Harbor.S3.SecretKey)
		}
		if err := p.CreateOrUpdateSecret("harbor-registry", Namespace, registrySecret); err != nil {
			return err
		}

		if err := p.CreateOrUpdateSecret("harbor-jobservice", Namespace, map[string][]byte{
			"secret": []byte(nonce),
		}); err != nil {
			return err
		}

		if err := p.CreateOrUpdateConfigMap("trusted-certs", Namespace, map[string]string{
			"ca.crt": string(p.GetIngressCA().GetPublicChain()[0].EncodedCertificate()),
		}); err != nil {
			return err
		}
		exporterSecret := map[string][]byte{
			"HARBOR_ADMIN_PASSWORD":    []byte(p.Harbor.AdminPassword),
			"HARBOR_DATABASE_PASSWORD": []byte(p.Harbor.DB.Password),
		}
		if err := p.CreateOrUpdateSecret("harbor-exporter", Namespace, exporterSecret); err != nil {
			return err
		}
	}

	for _, spec := range manifests {
		if err := p.ApplySpecs("", fmt.Sprintf("harbor/%s.yaml", spec)); err != nil {
			return err
		}
	}

	// Skip connecting if in dry run mode
	if p.ApplyDryRun {
		return nil
	}

	client, err := NewClient(p)
	if err != nil {
		return err
	}
	if err := client.UpdateSettings(*p.Harbor.Settings); err != nil {
		p.Errorf("Failed  to update harbor settings: %v", err)
	}
	return nil
}

func getHtPasswd(password string) ([]byte, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s:%s", HaborRegistryUsername, string(passwordBytes))), nil
}
