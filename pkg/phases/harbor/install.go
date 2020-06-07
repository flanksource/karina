package harbor

import (
	"fmt"
	"strings"

	pgapi "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/konfigadm/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	HaborRegistryUsername = "harbor_registry_user"
)

func Deploy(p *platform.Platform) error {
	if p.Harbor == nil || p.Harbor.Disabled {
		if err := p.DeleteSpecs("", "harbor.yaml", "harbor"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	p.Infof("Deploying harbor %s", p.Harbor.Version)

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
		var tokenKey []byte
		if harborCore != nil && (*harborCore)["tls.key"] != nil {
			// migrate key over from existing installation
			tokenKey = (*harborCore)["tls.key"]
		} else {
			tokenKey = p.NewSelfSigned("registry-token").EncodedPrivateKey()
		}
		if err := p.CreateOrUpdateSecret("token-key", Namespace, map[string][]byte{
			"tls.key": tokenKey,
		}); err != nil {
			return err
		}
	}

	if p.Harbor.DB == nil {
		dbConfig := pgapi.NewClusterConfig(dbCluster, dbNames...)
		db, err := postgresoperator.GetOrCreateDB(p, dbConfig)
		if err != nil {
			return fmt.Errorf("deploy: failed to get/update db: %v", err)
		}
		p.Harbor.DB = db
	}

	if err := p.CreateOrUpdateSecret("harbor-chartmuseum", Namespace, map[string][]byte{
		"CACHE_REDIS_PASSWORD":  []byte{},
		"AWS_SECRET_ACCESS_KEY": []byte(p.S3.SecretKey),
	}); err != nil {
		return err
	}

	if err := p.CreateOrUpdateSecret("harbor-clair", Namespace, map[string][]byte{
		"config.yaml": []byte(getClairConfig(p)),
		"database":    []byte(p.Harbor.DB.GetConnectionURL("clair")),
		"redis":       []byte("redis://harbor-redis:6379/4"),
	}); err != nil {
		return err
	}

	if err := p.CreateOrUpdateSecret("harbor-core", Namespace, map[string][]byte{
		"HARBOR_ADMIN_PASSWORD":        []byte(p.Harbor.AdminPassword),
		"POSTGRESQL_PASSWORD":          []byte(p.Harbor.DB.Password),
		"CLAIR_DB_PASSWORD":            []byte(p.Harbor.DB.Password),
		"REGISTRY_CREDENTIAL_PASSWORD": registryPassword,
		"secretKey":                    []byte("not-a-secure-key"),
		"secret":                       []byte(nonce),
		"CSRF_KEY":                     csrfKey,
	}); err != nil {
		return err
	}

	if err := p.CreateOrUpdateSecret("harbor-registry", Namespace, map[string][]byte{
		"REGISTRY_HTPASSWD":             registryHtPassword,
		"REGISTRY_HTTP_SECRET":          []byte(nonce),
		"REGISTRY_REDIS_PASSWORD":       []byte(""),
		"REGISTRY_STORAGE_S3_ACCESSKEY": []byte(p.S3.AccessKey),
		"REGISTRY_STORAGE_S3_SECRETKEY": []byte(p.S3.SecretKey),
	}); err != nil {
		return err
	}

	if err := p.CreateOrUpdateSecret("harbor-jobservice", Namespace, map[string][]byte{
		"secret": []byte(nonce),
	}); err != nil {
		return err
	}

	if err := p.ApplySpecs(Namespace, "harbor.yaml"); err != nil {
		return err
	}

	if err := p.ApplySpecs(Namespace, "harbor-exporter.yaml"); err != nil {
		return err
	}

	client, err := NewClient(p)
	if err != nil {
		return err
	}
	return client.UpdateSettings(*p.Harbor.Settings)
}

func getHtPasswd(password string) ([]byte, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s:%s", HaborRegistryUsername, string(passwordBytes))), nil
}

func getClairConfig(p *platform.Platform) string {
	return strings.ReplaceAll(fmt.Sprintf(`
clair:
  database:
    type: pgsql
    options:
      source: "%s"
      # Number of elements kept in the cache
      # Values unlikely to change (e.g. namespaces) are cached in order to save prevent needless roundtrips to the database.
      cachesize: 16384
  api:
    # API server port
    port: 6060
    healthport: 6061
    # Deadline before an API request will respond with a 503
    timeout: 300s
  updater:
    interval: 12h
	`, p.Harbor.DB.GetConnectionURL("clair")), "\t", "  ")
}
