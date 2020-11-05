package consul

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

type BackupRestore struct {
	*platform.Platform
	Name        string
	Namespace   string
	dockerImage string
}

func NewBackupRestore(platform *platform.Platform, name, namespace string) *BackupRestore {
	dockerImage := platform.Vault.Consul.BackupImage
	if dockerImage == "" {
		dockerImage = "docker.io/flanksource/consul-backup-restore:0.1"
	}
	br := &BackupRestore{
		Platform:    platform,
		Name:        name,
		Namespace:   namespace,
		dockerImage: dockerImage,
	}
	return br
}

func (b *BackupRestore) Backup() error {
	job := b.GenerateBackupJob().
		Command("/scripts/backup.sh").
		AsOneShotJob()

	if err := b.Apply(b.Namespace, job); err != nil {
		return err
	}

	return b.StreamLogs(b.Namespace, job.Name)
}

func (b *BackupRestore) ScheduleBackup(schedule string) error {
	job := b.GenerateBackupJob().
		Command("/scripts/backup.sh").
		AsCronJob(schedule)
	return b.Apply(b.Namespace, job)
}

func (b *BackupRestore) Restore(backup string) error {
	var backupBucket string
	if !strings.HasPrefix(backup, "s3://") {
		backupBucket = b.Vault.Consul.Bucket
		backup = fmt.Sprintf("s3://%s/consul/backups/%s/%s/%s", b.Vault.Consul.Bucket, b.Namespace, b.Name, backup)
	} else {
		uri, err := url.Parse(backup)
		if err != nil {
			return errors.Wrapf(err, "failed to parse s3 url %s", backup)
		}
		backupBucket = uri.Host
	}
	job := b.GenerateBackupJob().
		Command("/scripts/restore.sh").
		EnvVars(map[string]string{
			"RESTORE_URL":    backup,
			"RESTORE_BUCKET": backupBucket,
		}).AsOneShotJob()

	if err := b.Apply(b.Namespace, job); err != nil {
		return err
	}

	return b.StreamLogs(b.Namespace, job.Name)
}

func (b *BackupRestore) GenerateBackupJob() *kommons.DeploymentBuilder {
	vault := b.Vault
	consulBackupSecret := "consul-backup-config"

	builder := kommons.Deployment("consul-backup-"+b.Name+"-"+utils.ShortTimestamp(), b.dockerImage)
	return builder.
		EnvVarFromField("POD_NAMESPACE", "metadata.namespace").
		EnvVarFromSecret("AWS_ACCESS_KEY_ID", consulBackupSecret, "AWS_ACCESS_KEY_ID").
		EnvVarFromSecret("AWS_SECRET_ACCESS_KEY", consulBackupSecret, "AWS_SECRET_ACCESS_KEY").
		EnvVarFromSecret("AWS_ENDPOINT", consulBackupSecret, "AWS_ENDPOINT").
		EnvVarFromSecret("AWS_S3_FORCE_PATH_STYLE", consulBackupSecret, "AWS_S3_FORCE_PATH_STYLE").
		EnvVars(map[string]string{
			"CONSUL_ADDR":   fmt.Sprintf("%s-0.%s.%s.svc:8500", b.Name, b.Name, b.Namespace),
			"BACKUP_BUCKET": vault.Consul.Bucket,
			"BACKUP_PATH":   fmt.Sprintf("consul/backups/%s/%s/", b.Namespace, b.Name),
		}).
		Labels(map[string]string{
			"application": "consul-backup",
			"name":        fmt.Sprintf("consul-backup-%s-%s", b.Name, utils.RandomString(6)),
		})
}
