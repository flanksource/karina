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
		platform.PostgresOperator.Version = "v1.3.4.flanksource.2"
	}
	if platform.PostgresOperator.SpiloImage == "" {
		platform.PostgresOperator.SpiloImage = "docker.io/flanksource/spilo:1.6-p2.flanksource"
	}
	if platform.PostgresOperator.BackupImage == "" {
		platform.PostgresOperator.BackupImage = "docker.io/tobernguyen/postgres-backups:latest-9"
	}
	if platform.PostgresOperator.BackupBucket == "" {
		platform.PostgresOperator.BackupBucket = "postgres-backups-" + platform.Name
	}

	if err := platform.GetOrCreateBucket(platform.PostgresOperator.BackupBucket); err != nil {
		platform.Warnf("Failed to get/create s3://%s %v", platform.PostgresOperator.BackupBucket, err)
	}

	if platform.PostgresOperator.BackupSchedule == "" {
		platform.PostgresOperator.BackupSchedule = "30 0 * * *"
	}

	if err := platform.CreateOrUpdateNamespace("postgres-operator", nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "postgres-operator.yaml", "postgres-exporter-config.yaml", "postgres-operator-config.yaml", "template/postgres-db.yaml.raw")
}
