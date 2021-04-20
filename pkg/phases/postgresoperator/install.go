package postgresoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "postgres-operator"
)

func Deploy(platform *platform.Platform) error {
	if platform.PostgresOperator.IsDisabled() {
		if err := platform.DeleteSpecs("", "postgres-operator.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if platform.PostgresOperator.Version == "" {
		platform.PostgresOperator.Version = "v1.6.2"
	}
	if platform.PostgresOperator.SpiloImage == "" {
		platform.PostgresOperator.SpiloImage = "docker.io/flanksource/spilo:1.6-p2.flanksource"
	}
	if platform.PostgresOperator.BackupImage == "" {
		platform.PostgresOperator.BackupImage = "docker.io/flanksource/postgres-backups:v0.2.0"
	}
	if platform.PostgresOperator.DefaultBackupBucket == "" {
		platform.PostgresOperator.DefaultBackupBucket = "postgres-backups-" + platform.Name
	}

	if err := platform.GetOrCreateBucket(platform.PostgresOperator.DefaultBackupBucket); err != nil {
		platform.Warnf("Failed to get/create s3://%s %v", platform.PostgresOperator.DefaultBackupBucket, err)
	}

	if platform.PostgresOperator.DefaultBackupSchedule == "" {
		platform.PostgresOperator.DefaultBackupSchedule = "30 0 * * *"
	}

	if err := platform.CreateOrUpdateNamespace("postgres-operator", nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "postgres-operator.yaml", "postgres-operator-monitoring.yaml", "postgres-exporter-config.yaml", "postgres-operator-config.yaml", "template/postgresql-db.yaml.raw")
}
