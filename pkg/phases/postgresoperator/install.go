package postgresoperator

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
	if platform.PostgresOperator.Version == "" {
		platform.PostgresOperator.Version = "v1.3.4"
	}
	if platform.PostgresOperator.SpiloImage == "" {
		platform.PostgresOperator.SpiloImage = "docker.io/flanksource/spilo:1.6-p2.flanksource"
	}
	if platform.PostgresOperator.BackupImage == "" {
		platform.PostgresOperator.BackupImage = "docker.io/flanksource/postgres-backups:0.1.5"
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

	if err := platform.ApplySpecs(Namespace, "postgres-operator.crd.yaml"); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "postgres-operator-config.yaml"); err != nil {
		return err
	}
	return platform.ApplySpecs(Namespace, "postgres-operator.yaml")
}
