package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/flanksource/commons/text"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/kommons/proxy"
	"github.com/jackc/pgx/v4"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	dbv2 "github.com/flanksource/template-operator-library/api/db/v2"
)

const Namespace = "postgres-operator"

// nolint: revive
type PostgresDB struct {
	Name         string
	Namespace    string
	Superuser    string
	Secret       string
	version      string
	backupConfig *map[string][]byte
	client       *kommons.Client
}

type PostgresqlDB struct {
	*PostgresDB
	Restic       bool
	BackupBucket string
	platform     *platform.Platform
}

type s3Backup struct {
	Name string
	Date time.Time
}

func GetPostgresDB(client *kommons.Client, name string, resticEnabled bool) (*PostgresDB, error) {
	db := PostgresDB{client: client}

	_db := &api.Postgresql{TypeMeta: metav1.TypeMeta{
		Kind:       "postgresql",
		APIVersion: "acid.zalan.do",
	}}

	name = clusterName(name)

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

	if resticEnabled {
		backupConfig := db.client.GetSecret(Namespace, fmt.Sprintf("backup-%s-config", name))
		if backupConfig == nil {
			return nil, fmt.Errorf("failed to get backup config of %s", name)
		}
		db.backupConfig = backupConfig
	}

	db.Name = name
	db.Namespace = _db.Namespace
	db.version = _db.Spec.PgVersion
	db.Superuser = op.Configuration.PostgresUsersConfiguration.SuperUsername
	if db.Superuser == "" {
		db.Superuser = "postgres"
	}
	db.Secret = fmt.Sprintf("%s.%s.credentials", db.Superuser, name)
	return &db, nil
}

func GetPostgresqlDB(p *platform.Platform, name string) (*PostgresqlDB, error) {
	client := &p.Client
	_db := &dbv2.PostgresqlDB{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgresqlDB",
			APIVersion: "db.flanksource.com/v2",
		},
	}

	if err := client.Get(Namespace, name, _db); err != nil {
		return nil, fmt.Errorf("could not get db %v", err)
	}

	resticEnabled := _db.Spec.Backup.Restic
	postgresDB, err := GetPostgresDB(client, name, resticEnabled)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get postgres db")
	}

	db := PostgresqlDB{PostgresDB: postgresDB, platform: p}

	db.Restic = resticEnabled
	db.BackupBucket = _db.Spec.Backup.Bucket

	return &db, nil
}

func clusterName(name string) string {
	if !strings.HasPrefix(name, "postgres-") {
		return "postgres-" + name
	}
	return name
}

func (db *PostgresDB) GetClusterName() string {
	return clusterName(db.Name)
}

func (db *PostgresDB) OpenDB(database string) (*pgx.Conn, error) {
	name := db.GetClusterName()
	pod, err := db.client.WaitForPodByLabel(db.Namespace, fmt.Sprintf("cluster-name=%s,spilo-role=master", name), 30*time.Second)
	if err != nil {
		return nil, err
	}

	secretName := fmt.Sprintf("postgres.%s.credentials", name)
	secret := db.client.GetSecret("postgres-operator", secretName)
	if secret == nil {
		return nil, fmt.Errorf("%s not found", secretName)
	}
	db.client.Debugf("[%s] connecting with %s", pod.Name, string((*secret)["username"]))

	dialer, err := db.client.GetProxyDialer(proxy.Proxy{
		Namespace:    db.Namespace,
		Kind:         "pods",
		ResourceName: pod.Name,
		Port:         5432,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get proxy dialer")
	}

	cfg, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	cfg.Host = "127.0.0.1"
	cfg.Port = 5432
	cfg.Database = database
	cfg.User = string((*secret)["username"])
	cfg.Password = string((*secret)["password"])
	cfg.DialFunc = dialer.DialContext

	return pgx.ConnectConfig(context.Background(), cfg)
}

func (db *PostgresDB) WithConnection(database string, fn func(*pgx.Conn) error) error {
	pgdb, err := db.OpenDB(database)
	if err != nil {
		return err
	}
	defer pgdb.Close(context.Background())
	return fn(pgdb)
}

func (db *PostgresDB) String() string {
	return fmt.Sprintf("%s/%s[version=%s, secret=%s]", db.Namespace, db.Name, db.version, db.Secret)
}

func (db *PostgresDB) WaitFor(timeout time.Duration) error {
	_, err := db.client.WaitForResource("postgresql", db.Namespace, db.Name, timeout)
	return err
}

func (db *PostgresDB) Terminate() error {
	return db.client.DeleteByKind("postgresql", db.Namespace, db.Name)
}

func (db *PostgresqlDB) Backup() error {
	job := db.GenerateBackupJob().AsOneShotJob()

	if err := db.client.Apply(db.Namespace, job); err != nil {
		return err
	}

	jobPod, err := db.client.GetJobPod(Namespace, job.Name)
	if err != nil {
		return err
	}
	return db.client.StreamLogs(db.Namespace, jobPod)
}

func (db *PostgresDB) TriggerBackup(timeout time.Duration) error {
	backupCronJobName := fmt.Sprintf("backup-%s", db.GetClusterName())
	jobName, err := db.client.TriggerCronJobManually(db.Namespace, backupCronJobName)
	if err != nil {
		return errors.Wrap(err, "failed to trigger cronjob")
	}
	return db.client.WaitForJob(db.Namespace, jobName, timeout)
}

func (db *PostgresqlDB) ListBackups(limit int, quiet bool) ([]string, error) {
	if db.Restic {
		return db.ListResticBackups(limit, quiet)
	}

	return db.ListS3Backups(limit, quiet)
}

func (db *PostgresqlDB) ListS3Backups(limit int, quiet bool) ([]string, error) {
	s3Bucket := db.BackupBucket

	mc, err := db.platform.GetS3Client()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get s3 client")
	}
	ok, err := mc.BucketExists(s3Bucket)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query if bucket %s exists", s3Bucket)
	} else if !ok {
		return nil, errors.Errorf("bucket %s does not exist", s3Bucket)
	}

	prefix := fmt.Sprintf("%s/", clusterName(db.Name))
	doneCh := make(chan struct{})
	defer close(doneCh)
	obj := mc.ListObjectsV2(s3Bucket, prefix, false, doneCh)

	results := []s3Backup{}

	for o := range obj {
		date, err := getBackupDate(o.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse backup date")
		}
		results = append(results, s3Backup{Name: o.Key, Date: *date})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	count := len(results)
	if limit < count {
		count = limit
	}

	sPrintS3BackupPath := func(snapshot s3Backup) string {
		return fmt.Sprintf("s3://%s/%s", s3Bucket, snapshot.Name)
	}

	var backupPaths []string
	if quiet {
		for i := 0; i < count; i++ {
			backupPath := sPrintS3BackupPath(results[i])
			backupPaths = append(backupPaths, backupPath)
			fmt.Println(backupPath)
		}
	} else {
		w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "BACKUP PATH\tTIME\tAGE")
		for i := 0; i < count; i++ {
			snapshot := results[i]
			backupPath := sPrintS3BackupPath(snapshot)
			backupPaths = append(backupPaths, backupPath)
			fmt.Fprintf(w, "%s\t%s\t%s\n", backupPath, snapshot.Date.Format("2006-01-02 15:04:05 -07 MST"), text.HumanizeDuration(time.Since(snapshot.Date)))
		}
	}

	return backupPaths, nil
}

func (db *PostgresqlDB) ListResticBackups(limit int, quiet bool) ([]string, error) {
	backupConfig := *db.backupConfig
	s3Bucket := db.BackupBucket

	var backupS3Bucket string
	if s3Bucket != "" {
		backupS3Bucket = s3Bucket
	} else {
		backupS3Bucket = string(backupConfig["BACKUP_S3_BUCKET"])
	}

	job := kommons.
		Deployment("list-backups-"+db.Name+"-"+utils.ShortTimestamp(), string(backupConfig["BACKUP_IMAGE"])).
		EnvFromSecret(fmt.Sprintf("backup-%s-config", db.Name)).
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvVars(map[string]string{
			"BACKUP_S3_BUCKET": backupS3Bucket,
			"PGHOST":           db.Name,
		}).
		EnvVars(map[string]string{
			"RESTIC_REPOSITORY": "s3:$(AWS_ENDPOINT_URL)/$(BACKUP_S3_BUCKET)",
		}).
		Command("restic", "snapshots", "--json").
		AsOneShotJob()

	if err := db.client.Apply(db.Namespace, job); err != nil {
		return nil, err
	}

	if err := db.client.WaitForJob(Namespace, job.Name, 1*time.Minute); err != nil {
		return nil, err
	}

	jobPod, err := db.client.GetJobPod(Namespace, job.Name)
	if err != nil {
		return nil, err
	}

	podLogs, err := db.client.GetPodLogs(Namespace, jobPod, "")
	if err != nil {
		return nil, err
	}
	var resticSnapshots []ResticSnapshot
	err = json.Unmarshal([]byte(podLogs), &resticSnapshots)
	if err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(resticSnapshots) {
		resticSnapshots = resticSnapshots[:limit]
	}

	// Sort snapshots in Time descending order (newer backups first)
	sort.Slice(resticSnapshots, func(i, j int) bool {
		return resticSnapshots[i].Time.After(resticSnapshots[j].Time)
	})

	sPrintBackupPath := func(snapshot ResticSnapshot) string {
		return fmt.Sprintf("restic:s3:%s/%s%s", string(backupConfig["AWS_ENDPOINT_URL"]), backupS3Bucket, snapshot.Paths[0])
	}

	var backupPaths []string
	if quiet {
		for i := 0; i < len(resticSnapshots); i++ {
			backupPath := sPrintBackupPath(resticSnapshots[i])
			backupPaths = append(backupPaths, backupPath)
			fmt.Println(backupPath)
		}
	} else {
		w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "BACKUP PATH\tTIME\tAGE")
		for i := 0; i < len(resticSnapshots); i++ {
			snapshot := resticSnapshots[i]
			backupPath := sPrintBackupPath(snapshot)
			backupPaths = append(backupPaths, backupPath)
			fmt.Fprintf(w, "%s\t%s\t%s\n", backupPath, snapshot.Time.Format("2006-01-02 15:04:05 -07 MST"), text.HumanizeDuration(time.Since(snapshot.Time)))
		}
	}

	return backupPaths, nil
}

// Restore executes a job to retrieve the logical backup specified and then applies it to the database
// if trace is true, restore commands are printed to stdout
func (db *PostgresqlDB) Restore(fullBackupPath string, trace bool) error {
	if !strings.HasPrefix(fullBackupPath, "restic:s3:") {
		return db.RestoreS3(fullBackupPath, trace)
	}

	return db.RestoreRestic(fullBackupPath, trace)
}

func (db *PostgresqlDB) RestoreS3(backup string, trace bool) error {
	if !strings.HasPrefix(backup, "s3://") {
		backup = fmt.Sprintf("s3://%s/%s/%s", db.BackupBucket, db.Name, backup)
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

	if err := db.client.WaitForJob(db.Namespace, job.Name, 1*time.Minute); err != nil {
		return err
	}

	jobPod, err := db.client.GetJobPod(db.Namespace, job.Name)
	if err != nil {
		return err
	}

	return db.client.StreamLogs(db.Namespace, jobPod)
}

func (db *PostgresqlDB) RestoreRestic(fullBackupPath string, trace bool) error {
	db.client.Infof("[%s] restoring from backup %s", db.Name, fullBackupPath)

	re := regexp.MustCompile(`(s3:(s3.amazonaws.com|https?://[^/]+)/[^/]+)(/.+)`)
	matches := re.FindStringSubmatch(fullBackupPath)
	resticRepository := matches[1]
	backupPath := matches[3]
	psqlOpts := ""
	if trace {
		psqlOpts = "--echo-all"
	}
	jobBuilder := db.GenerateBackupJob().
		Command("/restore.sh").
		EnvVars(map[string]string{
			"BACKUP_PATH":      backupPath,
			"PSQL_BEFORE_HOOK": "",
			"PSQL_AFTER_HOOK":  "",
			"PSQL_OPTS":        psqlOpts,
		})
	jobBuilder.Name = "restore-" + db.Name + "-" + utils.ShortTimestamp()

	if resticRepository != "" {
		jobBuilder.EnvVars(map[string]string{
			"RESTIC_REPOSITORY": resticRepository,
		})
	}

	job := jobBuilder.AsOneShotJob()

	if err := db.client.Apply(db.Namespace, job); err != nil {
		return err
	}

	jobPod, err := db.client.GetJobPod(Namespace, job.Name)
	if err != nil {
		return err
	}
	return db.client.StreamLogs(db.Namespace, jobPod)
}

func (db *PostgresqlDB) GenerateBackupJob() *kommons.DeploymentBuilder {
	if db.Restic {
		return db.generateResticBackupJob()
	}

	return db.generateS3BackupJob()
}

func (db *PostgresqlDB) generateS3BackupJob() *kommons.DeploymentBuilder {
	backupImage := "docker.io/flanksource/postgres-backups:0.1.5"
	operatorSecret := "postgres-operator-cluster-environment"
	var builder = kommons.Deployment("backup-"+db.Name+"-"+utils.ShortTimestamp(), backupImage)

	return builder.
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvVarFromSecret("AWS_SECRET_ACCESS_KEY", operatorSecret, "AWS_SECRET_ACCESS_KEY").
		EnvVarFromSecret("AWS_ACCESS_KEY_ID", operatorSecret, "AWS_ACCESS_KEY_ID").
		EnvVarFromSecret("LOGICAL_BACKUP_S3_ENDPOINT", operatorSecret, "AWS_ENDPOINT_URL").
		EnvVarFromSecret("AWS_S3_FORCE_PATH_STYLE", operatorSecret, "AWS_S3_FORCE_PATH_STYLE").
		EnvVarFromSecret("LOGICAL_BACKUP_S3_REGION", operatorSecret, "AWS_REGION").
		EnvVarFromField("POD_NAMESPACE", "metadata.namespace").
		EnvVars(map[string]string{
			"PGHOST":                   db.Name,
			"PGPORT":                   "5432",
			"SCOPE":                    clusterName(db.Name),
			"PGSSLMODE":                "prefer",
			"PGDATABASE":               "postgres",
			"LOGICAL_BACKUP_S3_BUCKET": db.BackupBucket,
			"LOGICAL_BACKUP_S3_SSE":    "AES256",
			"PG_VERSION":               "12",
			"PGUSER":                   "postgres",
			"CLUSTER_NAME_LABEL":       "cluster-name",
		}).
		Resources(v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("500m"),
				v1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10m"),
				v1.ResourceMemory: resource.MustParse("128Mi"),
			},
		}).
		ServiceAccount("postgres-pod")
}

func (db *PostgresDB) generateResticBackupJob() *kommons.DeploymentBuilder {
	backupConfig := *db.backupConfig

	var builder = kommons.Deployment("backup-"+db.Name+"-"+utils.ShortTimestamp(), string(backupConfig["BACKUP_IMAGE"]))
	return builder.
		EnvVarFromSecret("PGPASSWORD", db.Secret, "password").
		EnvFromSecret(fmt.Sprintf("backup-%s-config", db.Name)).
		EnvVars(map[string]string{
			"PGHOST":            db.Name,
			"PGPORT":            "5432",
			"PGSSLMODE":         "prefer",
			"PGDATABASE":        "postgres",
			"PGUSER":            db.Superuser,
			"PG_VERSION":        db.version,
			"RESTIC_REPOSITORY": "s3:$(AWS_ENDPOINT_URL)/$(BACKUP_S3_BUCKET)",
		}).
		Labels(map[string]string{
			"application": "postgres-logical-backup",
		})
}

func getBackupDate(key string) (*time.Time, error) {
	filename := filepath.Base(key)
	if !strings.HasSuffix(filename, ".sql.gz") {
		return nil, errors.Errorf("expected filename %s to have .sql.gz extension", filename)
	}

	date, err := time.Parse("2006-01-02.150405.sql.gz", filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse date %s", filename)
	}
	return &date, nil
}
