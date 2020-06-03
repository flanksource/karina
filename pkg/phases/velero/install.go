package velero

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/karina/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace = "velero"
)

func Install(platform *platform.Platform) error {
	if platform.Velero == nil || platform.Velero.Disabled {
		if err := platform.DeleteSpecs(Namespace, "velero.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if platform.Velero.Version == "" {
		platform.Velero.Version = "v1.3.2"
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := platform.GetOrCreateBucket(platform.Velero.Bucket); err != nil {
		return err
	}

	if err := platform.CreateOrUpdateSecret("cloud-credentials", Namespace, map[string][]byte{
		"cloud": []byte(fmt.Sprintf(`[default]
	aws_access_key_id=%s
	aws_secret_access_key=%s`, platform.S3.AccessKey, platform.S3.SecretKey)),
	}); err != nil {
		return err
	}

	if platform.Velero.Config == nil {
		platform.Velero.Config = make(map[string]string)
	}

	endpoint := platform.S3.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}
	platform.Velero.Config["s3Url"] = endpoint
	platform.Velero.Config["region"] = platform.S3.Region
	platform.Velero.Config["insecureSkipTLSVerify"] = "true"
	platform.Velero.Config["s3ForcePathStyle"] = "true"

	return platform.ApplySpecs(Namespace, "velero.yaml")
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
			TTL:                metav1.Duration{Duration: time.Duration(30) * 24 * time.Hour},
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

	platform.Infof("Waiting for %s to complete", backup.Metadata.Name)
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
		}
		time.Sleep(5 * time.Second)
	}
}
