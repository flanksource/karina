package postgres

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/minio/minio-go/v6"

	"github.com/flanksource/commons/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/kommons"
)

const Namespace = "postgres-operator"

// nolint: golint
type PostgresDB struct {
	Name         string
	Namespace    string
	Secret       string
	version      string
	Superuser    string
	backupConfig *map[string][]byte
	client       *kommons.Client
}

type ResticSnapshot struct {
	Time    time.Time
	Paths   []string
	Tags    []string
	ID      string
	ShortID string `json:"short_id"`
}

func GetPostgresDB(client *kommons.Client, s3 *minio.Client, name string) (*PostgresDB, error) {
	db := PostgresDB{client: client}

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

	backupConfig := db.client.GetSecret(Namespace, fmt.Sprintf("backup-%s-config", name))
	if backupConfig == nil {
		return nil, fmt.Errorf("failed to get backup config of %s", name)
	}

	db.Name = name
	db.backupConfig = backupConfig
	db.Namespace = _db.Namespace
	db.version = _db.Spec.PgVersion
	db.Superuser = op.Configuration.PostgresUsersConfiguration.SuperUsername
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

func (db *PostgresDB) ListBackups(s3Bucket string, limit int, quiet bool) error {
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
		return err
	}

	jobPod, err := db.client.GetJobPod(Namespace, job.Name)
	if err != nil {
		return err
	}

	err = db.client.WaitForPod(Namespace, jobPod, 30*time.Second, v1.PodSucceeded)
	if err != nil {
		return err
	}

	podLogs, err := db.client.GetPodLogs(Namespace, jobPod, "")
	if err != nil {
		return err
	}
	var resticSnapshots []ResticSnapshot
	err = json.Unmarshal([]byte(podLogs), &resticSnapshots)
	if err != nil {
		return err
	}

	if limit > 0 {
		resticSnapshots = resticSnapshots[:limit]
	}

	// Sort snapshots in Time descending order (newer backups first)
	sort.Slice(resticSnapshots, func(i, j int) bool {
		return resticSnapshots[i].Time.After(resticSnapshots[j].Time)
	})

	if quiet {
		for i := 0; i < len(resticSnapshots); i++ {
			fmt.Println(resticSnapshots[i].Paths[0])
		}
	} else {
		w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "BACKUP PATH\tTIME\tAGE")
		for i := 0; i < len(resticSnapshots); i++ {
			snapshot := resticSnapshots[i]
			fmt.Fprintf(w, "%s\t%s\t%s\n", snapshot.Paths[0], snapshot.Time.Format("2006-01-01 15:04:05 -07 MST"), time.Since(snapshot.Time))
		}
	}

	return nil
}

func (db *PostgresDB) Restore(args ...string) error {
	backupConfig := *db.backupConfig

	var backupPath string
	var resticRepository string
	if len(args) == 2 {
		resticRepository = fmt.Sprintf("s3:%s/%s", string(backupConfig["AWS_ENDPOINT_URL"]), args[0])
		backupPath = args[1]
	} else {
		backupPath = args[0]
	}

	jobBuilder := db.GenerateBackupJob().
		Command("/restore.sh").
		EnvVars(map[string]string{
			"BACKUP_PATH":      backupPath,
			"PSQL_BEFORE_HOOK": "",
			"PSQL_AFTER_HOOK":  "",
			"PSQL_OPTS":        "--echo-all",
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
