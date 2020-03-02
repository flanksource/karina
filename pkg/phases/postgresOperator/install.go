package postgresOperator

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "postgres-operator"
)

func Deploy(platform *platform.Platform) error {
	if platform.PostgresOperator == nil || platform.PostgresOperator.Disabled {
		log.Infof("Postgres operator is disabled")
		return nil
	}

	if platform.PostgresOperator.BackupBucket == "" {
		platform.PostgresOperator.BackupBucket = "postgres-backups-" + platform.Name
	}

	if err := platform.GetOrCreateBucket(platform.PostgresOperator.BackupBucket); err != nil {
		return err
	}

	if platform.PostgresOperator.BackupSchedule == "" {
		platform.PostgresOperator.BackupSchedule = "30 0 * * *"
	}

	if err := platform.CreateOrUpdateNamespace("postgres-operator", nil, nil); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "postgres-operator.crd.yml"); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "postgres-operator-config.yml"); err != nil {
		return err
	}
	return platform.ApplySpecs(Namespace, "postgres-operator.yml")
}
