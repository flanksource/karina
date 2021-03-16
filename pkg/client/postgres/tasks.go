package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/utils"
	minio "github.com/minio/minio-go/v6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/kommons"
)

const Namespace = "postgres-operator"
const OperatorConfig = "default"

// nolint: golint
type PostgresDB struct {
	Name      string
	Namespace string
	Secret    string
	version   string
	Superuser string
	op        *api.OperatorConfiguration
	client    *kommons.Client
	s3        *minio.Client
}

func GetGenericPostgresDB(client *kommons.Client, s3 *minio.Client, namespace, name, secret, version string) (*PostgresDB, error) {
	db := PostgresDB{client: client, s3: s3}

	op := api.OperatorConfiguration{TypeMeta: metav1.TypeMeta{
		Kind:       "operatorconfiguration",
		APIVersion: "acid.zalan.do/v1",
	}}

	if err := client.Get(Namespace, OperatorConfig, &op); err != nil {
		return nil, fmt.Errorf("could not get opconfig %v", err)
	}

	db.op = &op
	db.Name = name
	db.Namespace = namespace
	db.Secret = secret
	db.version = version
	db.Superuser = "postgres"
	return &db, nil
}

func GetPostgresDB(client *kommons.Client, s3 *minio.Client, name string) (*PostgresDB, error) {
	db := PostgresDB{client: client, s3: s3}

	_db := &api.Postgresql{TypeMeta: metav1.TypeMeta{
		Kind:       "postgresql",
		APIVersion: "acid.zalan.do",
	}}

	if err := client.Get(Namespace, name, _db); err != nil {
		return nil, fmt.Errorf("could not get db %v", err)
	}

	op := api.OperatorConfiguration{TypeMeta: metav1.TypeMeta{
		Kind:       "operatorconfiguration",
		APIVersion: "acid.zalan.do/v1",
	}}
	if err := client.Get(Namespace, "default", &op); err != nil {
		return nil, fmt.Errorf("could not get opconfig %v", err)
	}

	db.op = &op
	db.Name = name
	db.Namespace = _db.Namespace
	db.version = _db.Spec.PgVersion
	db.Superuser = db.op.Configuration.PostgresUsersConfiguration.SuperUsername
	db.Secret = fmt.Sprintf("%s.%s.credentials", db.Superuser, name)
	return &db, nil
}

func (db *PostgresDB) String() string {
	return fmt.Sprintf("%s/%s[version=%s, secret=%s]", db.Namespace, db.Name, db.version, db.Secret)
}

func (db *PostgresDB) Backup() error {
	job := db.GenerateBackupJob().AsOneShotJob()

	if err := db.client.Apply(db.Namespace, job); err != nil {
		return err
	}

	return db.client.StreamLogs(db.Namespace, job.Name)
}

func (db *PostgresDB) ScheduleBackup(schedule string) error {
	job := db.GenerateBackupJob().AsCronJob(schedule)
	return db.client.Apply(db.Namespace, job)
}

type BackupItem struct {
	URL          string
	LastModified time.Time
	Size         int64
}

func (db *PostgresDB) ListBackups() ([]*BackupItem, error) {
	op := db.op.Configuration
	doneCh := make(chan struct{})
	defer close(doneCh)

	list := make([]*BackupItem, 0)

	for o := range db.s3.ListObjectsV2(op.LogicalBackup.S3Bucket, db.Name, true, doneCh) {
		url := fmt.Sprintf("s3://%s/%s", op.LogicalBackup.S3Bucket, o.Key)
		bi := &BackupItem{URL: url, LastModified: o.LastModified, Size: o.Size}
		list = append(list, bi)
	}

	return list, nil
}

func (db *PostgresDB) Restore(backup string) error {
	if !strings.HasPrefix(backup, "s3://") {
		backup = fmt.Sprintf("s3://%s/%s", db.op.Configuration.LogicalBackup.S3Bucket, backup)
	}
	job := db.GenerateBackupJob().
		Command("/restore.sh").
		EnvVars(map[string]string{
			"PATH_TO_BACKUP":   backup,
			"PSQL_BEFORE_HOOK": "",
			"PSQL_AFTER_HOOK":  "",
			"PGHOST":           db.Name,
			"PSQL_OPTS":        "--echo-all",
		}).AsOneShotJob()

	if err := db.client.Apply(db.Namespace, job); err != nil {
		return err
	}

	return db.client.StreamLogs(db.Namespace, job.Name)
}

func (db *PostgresDB) GenerateBackupJob() *kommons.DeploymentBuilder {
	op := db.op.Configuration

	builder := kommons.Deployment("backup-"+db.Name+"-"+utils.ShortTimestamp(), op.LogicalBackup.DockerImage)
	return builder.
		EnvVarFromField("POD_NAMESPACE", "metadata.namespace").
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvVars(map[string]string{
			"SCOPE":                      db.Name,
			"CLUSTER_NAME_LABEL":         op.Kubernetes.ClusterNameLabel,
			"LOGICAL_BACKUP_S3_BUCKET":   op.LogicalBackup.S3Bucket,
			"LOGICAL_BACKUP_S3_REGION":   op.LogicalBackup.S3Region,
			"LOGICAL_BACKUP_S3_ENDPOINT": op.LogicalBackup.S3Endpoint,
			"LOGICAL_BACKUP_S3_SSE":      op.LogicalBackup.S3SSE,
			"PGHOST":                     db.Name,
			"PG_VERSION":                 db.version,
			"PGPORT":                     "5432",
			"PGUSER":                     db.Superuser,
			"AWS_ACCESS_KEY_ID":          op.LogicalBackup.S3AccessKeyID,
			"AWS_SECRET_ACCESS_KEY":      op.LogicalBackup.S3SecretAccessKey,
			"PGDATABASE":                 "postgres",
			"PGSSLMODE":                  "prefer",
		}).
		Labels(map[string]string{
			op.Kubernetes.ClusterNameLabel: db.Name,
			"application":                  "spilo-logical-backup",
		}).
		Annotations(op.Kubernetes.CustomPodAnnotations).
		// Annotations(db.Spec.PodAnnotations).
		ServiceAccount(op.Kubernetes.PodServiceAccountName).
		Ports(8080, 5432, 8008)
}
