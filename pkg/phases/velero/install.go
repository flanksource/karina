package velero

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "velero"
)

func Install(platform *platform.Platform) error {
	if platform.Velero == nil || platform.Velero.Disabled {
		return nil
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	s3Client, err := platform.GetS3Client()
	if err != nil {
		return fmt.Errorf("install: failed to get S3 client: %v", err)
	}

	exists, err := s3Client.BucketExists(platform.Velero.Bucket)
	if err != nil {
		return fmt.Errorf("install: failed to check S3 bucket: %v", err)
	}
	if !exists {
		if err := s3Client.MakeBucket(platform.Velero.Bucket, platform.S3.Region); err != nil {
			return fmt.Errorf("install: failed to create S3 bucket: %v", err)
		}
	}
	secret := text.ToFile(fmt.Sprintf(`[default]
aws_access_key_id=%s
aws_secret_access_key=%s`, platform.S3.AccessKey, platform.S3.SecretKey), "")

	defer os.Remove(secret)

	velero := platform.GetBinaryWithKubeConfig("velero")
	backupConfig := fmt.Sprintf("region=%s,insecureSkipTLSVerify=true,s3ForcePathStyle=\"true\",s3Url=%s", platform.S3.Region, platform.S3.Endpoint)

	if err := velero("install --provider aws --plugins velero/velero-plugin-for-aws:v1.0.0 --bucket %s --secret-file %s --backup-location-config %s", platform.Velero.Bucket, secret, backupConfig); err != nil {
		return fmt.Errorf("install: failed to install velero: %v", err)
	}
	return nil

}

func CreateBackup(platform *platform.Platform) (*Backup, error) {
	no := false
	name := "backup-" + time.Now().Format("20060102-150405")
	backup := &Backup{
		Metadata: metav1.ObjectMeta{
			Namespace: Namespace,
			Name:      name,
		},
		Spec: BackupSpec{
			IncludedNamespaces: []string{"*"},
			TTL:                metav1.Duration{time.Duration(30) * 24 * time.Hour},
			StorageLocation:    "default",
			SnapshotVolumes:    &no,
		},
	}
	backup.APIVersion = "velero.io/v1"
	backup.Kind = "Backup"
	err := platform.Apply(Namespace, backup)
	if err != nil {
		return nil, fmt.Errorf("createBackup: failed to apply velero backup %v", err)
	}
	start := time.Now()

	log.Infof("Waiting for %s to complete", backup.Metadata.Name)
	for {
		backup = &Backup{}
		if err := platform.Get(Namespace, name, backup); err != nil {
			return nil, fmt.Errorf("createBackup: failed to get velero backup %v", err)
		}
		if backup.Status.Phase == BackupPhaseCompleted {
			return backup, nil
		} else if backup.Status.Phase != "" && backup.Status.Phase != BackupPhaseInProgress && backup.Status.Phase != BackupPhaseNew {
			return backup, fmt.Errorf("backup did not complete successfully %s", backup.Status.Phase)
		}

		if time.Now().After(start.Add(5 * time.Minute)) {
			return nil, fmt.Errorf("timeout exceeded")
		} else {
			time.Sleep(5 * time.Second)
		}
	}

}
