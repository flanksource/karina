package postgresoperator

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/flanksource/kommons/api/v1"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	pgclient "github.com/flanksource/karina/pkg/client/postgres"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	postgresdbv2 "github.com/flanksource/template-operator-library/api/db/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Link struct {
	ID  int64
	URL string
}

func Test(p *platform.Platform, test *console.TestResults) {
	if p.PostgresOperator.IsDisabled() {
		return
	}
	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	if p.E2E {
		TestLogicalBackupE2E(p, test)
		// TODO: re-enable this test
		// t pugitTestCloneDBFromWAL(p, test)
	}
}

// TestLogicalBackupE2E will test the logical backup function that comes with
// the db.flanksource.com/PostgresqlDB by:
// - Create a PG Cluster with db.flanksource.com/PostgresqlDB
// - Insert test fixtures to the database
// - Trigger the backup CronJob so we can have a fresh logical backup of the cluster
// - Spin up a new PG Cluster with db.flanksource.com/PostgresqlDB
// - Run the restore command to restore the data of the first PG Cluster to the second cluster
// - Check the test fixtures is in the second cluster
func TestLogicalBackupE2E(p *platform.Platform, test *console.TestResults) {
	testName := "pg-operator-logical-backup-e2e"
	db1, err := newDB(p, Namespace, "cluster1")
	if err != nil {
		test.Failf(testName, "error creating %v", err)
		return
	}
	if !p.PlatformConfig.Trace {
		defer db1.Terminate() //nolint: errcheck
	}
	db2, err := newDB(p, Namespace, "cluster2")
	if err != nil {
		test.Failf(testName, "error creating %v", err)
		return
	}
	if !p.PlatformConfig.Trace {
		defer db2.Terminate() //nolint: errcheck
	}

	if err := db1.WithConnection("postgres", insertTestFixtures); err != nil {
		test.Failf(testName, "failed to insert fixtures into PG Cluster %s: %v", db1.Name, err)
		return
	}

	if err := db1.TriggerBackup(5 * time.Minute); err != nil {
		test.Failf(testName, "failed to trigger backup %s: %v", db1.Name, err)
		return
	}

	backupPaths, err := db1.ListBackups("", 1, true)
	if err != nil {
		test.Failf("failed to get list of backups for Postgresql Cluster %s: %v", db1.Name, err)
		return
	}
	if len(backupPaths) < 1 {
		test.Failf("there is no backup found for Postgresql Cluster %s. Expected at least 1.", db1.Name)
		return
	}

	if err := db2.Restore(backupPaths[0], false); err != nil {
		test.Failf("failed to restore backup %s to Postgresql Cluster %s: %v", backupPaths[0], db2.Name, err)
		return
	}

	if err := db2.WithConnection("postgres", testFixturesArePresent); err != nil {
		test.Failf(testName, "failed to find test fixtures data in PG Cluster %s: %v", db2.Name, err)
		return
	}
	test.Passf(testName, "Tested E2E Backup and restore successfully")
}

func newDB(p *platform.Platform, namespace, name string) (*pgclient.PostgresDB, error) {
	cluster := &postgresdbv2.PostgresqlDB{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgresqlDB",
			APIVersion: "db.flanksource.com/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("e2e-test-%s-%s", utils.RandomString(6), name),
			Namespace: namespace,
		},
		Spec: postgresdbv2.PostgresqlDBSpec{
			Storage: postgresdbv2.Storage{
				Size: "500Mi",
			},
			Backup: postgresdbv2.PostgresqlBackup{
				Bucket: fmt.Sprintf("logical-backup-test-%s", utils.RandomString(6)),
				Retention: postgresdbv2.BackupRetention{
					KeepLast: 5,
				},
			},
			Replicas: 1,
		},
		Status: postgresdbv2.PostgresqlDBStatus{
			Conditions: []v1.Condition{},
		},
	}

	if err := p.Apply(cluster.Namespace, cluster); err != nil {
		return nil, errors.Wrap(err, "error creating db")
	}
	if _, err := p.WaitFor(cluster, 5*time.Minute); err != nil {
		return nil, errors.Wrap(err, "failed waiting for postgres to come up")
	}
	return pgclient.GetPostgresDB(&p.Client, cluster.Name)
}

func insertTestFixtures(pg *pgx.Conn) error {
	_, err := pg.Exec(context.TODO(), `
	CREATE TABLE links ( name VARCHAR(50) );
	INSERT INTO links VALUES('a');
	INSERT INTO links VALUES('b');
	`)
	return err
}

func testFixturesArePresent(pg *pgx.Conn) error {
	count := 0
	err := pg.QueryRow(context.Background(), "SELECT count(*) from links").Scan(&count)
	if err != nil {
		return errors.Wrap(err, "failed to list links")
	} else if count != 2 {
		return fmt.Errorf("expected 2 rows, got %d", count)
	}
	return nil
}
