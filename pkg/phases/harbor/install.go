package harbor

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/konfigadm/pkg/utils"
	pgapi "github.com/moshloop/platform-cli/pkg/api/postgres"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Harbor == nil || p.Harbor.Disabled {
		log.Infof("Skipping deployment of harbor, it is disabled")
		return nil
	}
	log.Infof("Deploying harbor %s", p.Harbor.Version)

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

	if p.Harbor.DB == nil {
		dbConfig := pgapi.NewClusterConfig(dbCluster, dbNames...)
		db, err := p.GetOrCreateDB(dbConfig)
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

	coreCert, err := p.CreateIngressCertificate("harbor-core")
	if err != nil {
		return err
	}
	tls := coreCert.AsTLSSecret()

	if err := p.CreateOrUpdateSecret("harbor-core", Namespace, map[string][]byte{
		"HARBOR_ADMIN_PASSWORD": []byte(p.Harbor.AdminPassword),
		"POSTGRESQL_PASSWORD":   []byte(p.Harbor.DB.Password),
		"CLAIR_DB_PASSWORD":     []byte(p.Harbor.DB.Password),
		"tls.key":               tls["tls.key"],
		"tls.crt":               tls["tls.crt"],
		"ca.crt":                tls["tls.crt"],
		"secretKey":             []byte("not-a-secure-key"),
		"secret":                []byte(nonce),
	}); err != nil {
		return err
	}

	if err := p.CreateTLSSecret(Namespace, "harbor", "harbor-ingress"); err != nil {
		return err
	}

	if err := p.CreateOrUpdateSecret("harbor-registry", Namespace, map[string][]byte{
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
	client := NewClient(p)
	return client.UpdateSettings(*p.Harbor.Settings)
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
