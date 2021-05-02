package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/kommons"
)

const Namespace = "postgres-operator"

// nolint: golint
type PostgresDB struct {
	Name         string
	Namespace    string
	Superuser    string
	Secret       string
	version      string
	backupConfig *map[string][]byte
	client       *kommons.Client
}

func GetPostgresDB(client *kommons.Client, name string) (*PostgresDB, error) {
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

	backupConfig := db.client.GetSecret(Namespace, fmt.Sprintf("backup-%s-config", name))
	if backupConfig == nil {
		return nil, fmt.Errorf("failed to get backup config of %s", name)
	}

	db.Name = name
	db.backupConfig = backupConfig
	db.Namespace = _db.Namespace
	db.version = _db.Spec.PgVersion
	db.Superuser = op.Configuration.PostgresUsersConfiguration.SuperUsername
	if db.Superuser == "" {
		db.Superuser = "postgres"
	}
	db.Secret = fmt.Sprintf("%s.%s.credentials", db.Superuser, name)
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

func (db *PostgresDB) Backup() error {
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

func (db *PostgresDB) ScheduleBackup(schedule string) error {
	job := db.GenerateBackupJob().AsCronJob(schedule)
	return db.client.Apply(db.Namespace, job)
}

func (db *PostgresDB) TriggerBackup(timeout time.Duration) error {
	backupCronJobName := fmt.Sprintf("backup-%s", db.GetClusterName())
	jobName, err := db.client.TriggerCronJobManually(db.Namespace, backupCronJobName)
	if err != nil {
		return errors.Wrap(err, "failed to trigger cronjob")
	}
	return db.client.WaitForJob(db.Namespace, jobName, timeout)
}

func (db *PostgresDB) ListBackups(s3Bucket string, limit int, quiet bool) ([]string, error) {
	backupConfig := *db.backupConfig

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

	if limit > 0 {
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
func (db *PostgresDB) Restore(fullBackupPath string, trace bool) error {
	if !strings.HasPrefix(fullBackupPath, "restic:s3:") {
		return fmt.Errorf("backup path format is not supported. It should start with \"restic:s3:\"")
	}

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

func (db *PostgresDB) GenerateBackupJob() *kommons.DeploymentBuilder {
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
