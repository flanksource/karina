package postgres

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/minio/minio-go/v6"

	"github.com/flanksource/commons/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/kommons"
)

const Namespace = "postgres-operator"
const OperatorConfig = "default"

// nolint: golint
type PostgresDB struct {
	Name            string
	Namespace       string
	Secret          string
	version         string
	Superuser       string
	backupBucket    string
	backupRetention *map[string][]byte
	opClusterEnv    *map[string][]byte
	op              *api.OperatorConfiguration
	client          *kommons.Client
}

func GetGenericPostgresDB(client *kommons.Client, s3 *minio.Client, namespace, name, secret, version string) (*PostgresDB, error) {
	db := PostgresDB{client: client}

	op := api.OperatorConfiguration{TypeMeta: metav1.TypeMeta{
		Kind:       "operatorconfiguration",
		APIVersion: "acid.zalan.do/v1",
	}}

	if err := client.Get(Namespace, OperatorConfig, &op); err != nil {
		return nil, fmt.Errorf("could not get opconfig %v", err)
	}

	opClusterEnv := db.client.GetSecret(Namespace, "postgres-operator-cluster-environment")

	db.opClusterEnv = opClusterEnv
	db.backupRetention = opClusterEnv // Backup retention should be part of postgres-operator-cluster-environment secret
	db.backupBucket = string((*opClusterEnv)["LOGICAL_BACKUP_S3_BUCKET"])
	db.op = &op
	db.Name = name
	db.Namespace = namespace
	db.Secret = secret
	db.version = version
	db.Superuser = db.op.Configuration.PostgresUsersConfiguration.SuperUsername
	return &db, nil
}

func GetPostgresDB(client *kommons.Client, s3 *minio.Client, name string) (*PostgresDB, error) {
	db := PostgresDB{client: client}

	opClusterEnv := db.client.GetSecret(Namespace, "postgres-operator-cluster-environment")

	postgresDBName := strings.TrimPrefix(name, "postgres-")
	postgresqlDB, err := client.GetByKind("PostgresqlDB", Namespace, postgresDBName)
	if err != nil {
		return nil, fmt.Errorf("could not get PostgresqlDB/%s", postgresDBName)
	}

	defaultBackupBucket := string((*opClusterEnv)["LOGICAL_BACKUP_S3_BUCKET"])
	var backupBucket string
	if postgresqlDB == nil {
		// could not find flanksource/PostgresqlDB, fallback to the default backup bucket
		backupBucket = defaultBackupBucket
	} else if bucket, _, _ := unstructured.NestedString(postgresqlDB.Object, "spec", "backup", "bucket"); bucket == "" {
		// spec.backup.bucket is not defined in flanksource/PostgresqlDB, fallback to the default backup bucket
		backupBucket = defaultBackupBucket
	} else {
		backupBucket = bucket
	}

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

	backupRetention := db.client.GetSecret(Namespace, fmt.Sprintf("backup-%s-config", name))
	if backupRetention == nil {
		return nil, fmt.Errorf("failed to get backup config of %s", name)
	}

	db.opClusterEnv = db.client.GetSecret(Namespace, "postgres-operator-cluster-environment")
	db.backupRetention = backupRetention
	db.backupBucket = backupBucket
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

func (db *PostgresDB) ListBackups(s3Bucket string) error {
	opClusterEnv := *db.opClusterEnv

	var backupS3Bucket string
	if s3Bucket != "" {
		backupS3Bucket = s3Bucket
	} else {
		backupS3Bucket = db.backupBucket
	}

	job := kommons.
		Deployment("list-backups-"+db.Name+"-"+utils.ShortTimestamp(), string(opClusterEnv["BACKUP_IMAGE"])).
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvVars(map[string]string{
			"AWS_ACCESS_KEY_ID":     string(opClusterEnv["AWS_ACCESS_KEY_ID"]),
			"AWS_SECRET_ACCESS_KEY": string(opClusterEnv["AWS_SECRET_ACCESS_KEY"]),
			"AWS_DEFAULT_REGION":    string(opClusterEnv["AWS_DEFAULT_REGION"]),
			"RESTIC_REPOSITORY":     fmt.Sprintf("s3:%s/%s", string(opClusterEnv["AWS_ENDPOINT_URL"]), backupS3Bucket),
			"RESTIC_PASSWORD":       string(opClusterEnv["BACKUP_PASSWORD"]),
		}).
		Command("restic", "snapshots", "--tag", db.Name).
		AsOneShotJob()

	defer func() {
		if err := db.client.DeleteByKind("Pod", db.Namespace, job.Name); err != nil {
			fmt.Sprintf("Failed to clean up pod %s", job.Name)
		}
	}()
	if err := db.client.Apply(db.Namespace, job); err != nil {
		return err
	}
	return db.client.StreamLogs(db.Namespace, job.Name)
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
	opClusterEnv := *db.opClusterEnv
	backupRetention := *db.backupRetention

	var builder = kommons.Deployment("backup-"+db.Name+"-"+utils.ShortTimestamp(), string(opClusterEnv["BACKUP_IMAGE"]))
	return builder.
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvVars(map[string]string{
			"AWS_ACCESS_KEY_ID":             string(opClusterEnv["AWS_ACCESS_KEY_ID"]),
			"AWS_SECRET_ACCESS_KEY":         string(opClusterEnv["AWS_SECRET_ACCESS_KEY"]),
			"AWS_DEFAULT_REGION":            string(opClusterEnv["AWS_DEFAULT_REGION"]),
			"RESTIC_REPOSITORY":             fmt.Sprintf("s3:%s/%s", string(opClusterEnv["AWS_ENDPOINT_URL"]), db.backupBucket),
			"RESTIC_PASSWORD":               string(opClusterEnv["BACKUP_PASSWORD"]),
			"BACKUP_RETENTION_KEEP_LAST":    string(backupRetention["BACKUP_RETENTION_KEEP_LAST"]),
			"BACKUP_RETENTION_KEEP_HOURLY":  string(backupRetention["BACKUP_RETENTION_KEEP_HOURLY"]),
			"BACKUP_RETENTION_KEEP_DAILY":   string(backupRetention["BACKUP_RETENTION_KEEP_DAILY"]),
			"BACKUP_RETENTION_KEEP_WEEKLY":  string(backupRetention["BACKUP_RETENTION_KEEP_WEEKLY"]),
			"BACKUP_RETENTION_KEEP_MONTHLY": string(backupRetention["BACKUP_RETENTION_KEEP_MONTHLY"]),
			"BACKUP_RETENTION_KEEP_YEARLY":  string(backupRetention["BACKUP_RETENTION_KEEP_YEARLY"]),
			"PGHOST":                        db.Name,
			"PGPORT":                        "5432",
			"PGSSLMODE":                     "prefer",
			"PGDATABASE":                    "postgres",
			"PGUSER":                        db.Superuser,
			"PG_VERSION":                    db.version,
		}).
		Labels(map[string]string{
			op.Kubernetes.ClusterNameLabel: db.Name,
			"application":                  "postgres-logical-backup",
		})
}
