package postgresoperator

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/flanksource/kommons/api/v1"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	pgapi "github.com/flanksource/karina/pkg/api/postgres"
	pgclient "github.com/flanksource/karina/pkg/client/postgres"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	postgresdbv2 "github.com/flanksource/template-operator-library/api/db/v2"
	"github.com/go-pg/pg/v9/orm"
	"github.com/minio/minio-go/v6"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Link struct {
	ID  int64
	URL string
}

type ClusterResponse struct {
	Members []ClusterResponseMember `json:"members"`
}

type ClusterResponseMember struct {
	Name     string      `json:"name"`
	Host     string      `json:"host"`
	Port     int         `json:"port"`
	Role     string      `json:"role"`
	State    string      `json:"state"`
	URL      string      `json:"api_url"`
	Timeline int         `json:"timeline"`
	Lag      interface{} `json:"lag"`
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
	cluster1 := newPostgresqlDBCluster(Namespace, "cluster1")
	defer removeLogicalBackupE2ECluster(p, test, cluster1)
	if err := p.Apply(Namespace, cluster1); err != nil {
		test.Failf(testName, "error creating db %s: %v", cluster1.Name, err)
		return
	}
	cluster1ZalandoPsqlName := fmt.Sprintf("postgres-%s", cluster1.Name)
	if err := waitForPgClusterToReady(p, Namespace, cluster1ZalandoPsqlName, 5*time.Minute); err != nil {
		test.Failf("postgres cluster %s failed to start: %s", cluster1ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "Postgresql Cluster %s is created successfully", cluster1ZalandoPsqlName)

	if err := insertTestFixtures(p, cluster1ZalandoPsqlName, test); err != nil {
		test.Failf(testName, "failed to insert fixtures into PG Cluster %s: %v", cluster1ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "test fixtures has been inserted to %s", cluster1ZalandoPsqlName)

	backupCronJobName := fmt.Sprintf("backup-postgres-%s", cluster1.Name)
	jobName, err := p.TriggerCronJobManually(Namespace, backupCronJobName)
	if err != nil {
		test.Failf(testName, "failed to trigger the CronJob %s: %v", backupCronJobName, err)
		return
	}

	if err := p.WaitForJob(Namespace, jobName, 5*time.Minute); err != nil {
		test.Failf(testName, "failed to create a logical backup of pg cluster %s: %v", cluster1ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "created a logical backup for %s", cluster1ZalandoPsqlName)

	cluster2 := newPostgresqlDBCluster(Namespace, "cluster2")
	defer removeLogicalBackupE2ECluster(p, test, cluster2)
	if err := p.Apply(Namespace, cluster2); err != nil {
		test.Failf(testName, "error creating db %s: %v", cluster1.Name, err)
		return
	}
	cluster2ZalandoPsqlName := fmt.Sprintf("postgres-%s", cluster2.Name)
	if err := waitForPgClusterToReady(p, Namespace, cluster2ZalandoPsqlName, 5*time.Minute); err != nil {
		test.Failf("postgres cluster %s failed to start: %s", cluster2ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "Postgresql Cluster %s is created successfully", cluster2ZalandoPsqlName)

	db1, err := pgclient.GetPostgresDB(&p.Client, cluster1ZalandoPsqlName)
	if err != nil {
		test.Failf("failed to get Postgres Client for %s: %v", cluster1ZalandoPsqlName, err)
		return
	}
	backupPaths, err := db1.ListBackups("", 1, true)
	if err != nil {
		test.Failf("failed to get list of backups for Postgresql Cluster %s: %v", cluster1ZalandoPsqlName, err)
		return
	}
	if len(backupPaths) < 1 {
		test.Failf("there is no backup found for Postgresql Cluster %s. Expected at least 1.", cluster1ZalandoPsqlName)
		return
	}
	test.Passf(testName, "get the latest backup of %s", cluster1ZalandoPsqlName)
	db2, err := pgclient.GetPostgresDB(&p.Client, cluster2ZalandoPsqlName)
	if err != nil {
		test.Failf("failed to get Postgres Client for %s: %v", cluster2ZalandoPsqlName, err)
		return
	}
	backupPathToRestore := backupPaths[0]
	err = db2.Restore(backupPathToRestore)
	if err != nil {
		test.Failf("failed to restore backup %s to Postgresql Cluster %s: %v", backupPathToRestore, cluster2ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "restore logical backup \"%s\" of %s to %s successfully", backupPathToRestore, cluster1ZalandoPsqlName, cluster2ZalandoPsqlName)
	if err := testFixturesArePresent(p, cluster2ZalandoPsqlName, 5*time.Minute, test); err != nil {
		test.Failf(testName, "failed to find test fixtures data in PG Cluster %s: %v", cluster2ZalandoPsqlName, err)
		return
	}
	test.Passf(testName, "verified test fixtures is in %s after the restore", cluster2ZalandoPsqlName)
}

func newPostgresqlDBCluster(namespace, name string) *postgresdbv2.PostgresqlDB {
	return &postgresdbv2.PostgresqlDB{
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
}

// TODO: replace with platform.WaitForResource
func waitForPgClusterToReady(p *platform.Platform, namespace, postgresqlClusterName string, timeout time.Duration) error {
	start := time.Now()
	db := &pgapi.Postgresql{}
	for {
		_ = p.Get(namespace, postgresqlClusterName, db)
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("timeout exceeded waiting for db %s to be ready, current state: %s", db.Name, db.Status.PostgresClusterStatus)
		}
		if db.Status.PostgresClusterStatus == pgapi.ClusterStatusRunning {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
}

func removeLogicalBackupE2ECluster(p *platform.Platform, test *console.TestResults, db *postgresdbv2.PostgresqlDB) {
	client, _, _, err := p.GetDynamicClientFor(Namespace, db)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}

	_, err = client.Get(context.TODO(), db.Name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return
	}

	if err := client.Delete(context.TODO(), db.Name, metav1.DeleteOptions{}); err != nil {
		test.Warnf("Failed to delete resource %s/%s/%s in namespace %s", db.APIVersion, db.Kind, db.Name, Namespace)
		return
	}
	test.Infof("Deleted %s/%s/%s in namespace %s", db.APIVersion, db.Kind, db.Name, Namespace)
}

// func TestCloneDBFromWAL(p *platform.Platform, test *console.TestResults) {
// 	testName := "postgres-operator-e2e"
// 	cluster1 := pgapi.NewClusterConfig(utils.RandomString(6), "test", "e2e_db")
// 	cluster1.BackupSchedule = "*/1 * * * *"
// 	cluster1Name := "postgres-" + cluster1.Name
// 	_, err := GetOrCreateDB(p, cluster1)
// 	defer removeE2ECluster(p, cluster1, test) //failsafe removal of cluster
// 	if err != nil {
// 		test.Failf(testName, "Error creating db %s: %v", cluster1.Name, err)
// 		return
// 	}
// 	test.Passf(testName, "Cluster %s deployed", cluster1Name)

// 	if err := insertTestFixtures(p, cluster1Name, test); err != nil {
// 		test.Failf(testName, "Failed to insert fixtures into database %s: %v", cluster1.Name, err)
// 		return
// 	}
// 	timestamp := time.Now().Add(5 * time.Second)

// 	if err := waitForWalBackup(p, cluster1Name, 5*time.Minute, timestamp, test); err != nil {
// 		test.Failf(testName, "Failed to find any wal backups for database %s: %v", cluster1.Name, err)
// 		return
// 	}

// 	cluster2 := pgapi.NewClusterConfig(cluster1.Name+"-clone", "test")
// 	cluster2.Clone = &pgapi.CloneConfig{
// 		ClusterName: cluster1Name,
// 		Timestamp:   time.Now().Format("2006-01-02 15:04:05 UTC"),
// 	}
// 	cluster2Name := "postgres-" + cluster2.Name
// 	_, err = GetOrCreateDB(p, cluster2)
// 	defer removeE2ECluster(p, cluster2, test)
// 	if err != nil {
// 		test.Failf(testName, "Error creating db %s: %v", cluster2.Name, err)
// 		return
// 	}
// 	test.Passf(testName, "Cluster %s deployed user", cluster1.Name)

// 	if err := testFixturesArePresent(p, cluster2Name, 5*time.Minute, test); err != nil {
// 		test.Failf(testName, "Failed to find test fixtures data in clone database %s: %v", cluster2Name, err)
// 		return
// 	}

// 	var errMessage error = nil

// 	ok := doUntil(func() bool {
// 		errMessage = checkReplicaLag(p, cluster1Name, cluster2Name)
// 		return errMessage == nil
// 	})

// 	if ok {
// 		test.Passf(testName, "Cloned cluster %s successfully created", cluster2Name)
// 	} else {
// 		test.Failf(testName, "Failed to check replica lag: %v", errMessage)
// 	}
// }

func insertTestFixtures(p *platform.Platform, clusterName string, test *console.TestResults) error {
	pgdb, err := p.OpenDB(Namespace, clusterName, "postgres")
	if err != nil {
		test.Failf("postgres-operator", "failed to connect to postgres")
		return err
	}
	defer pgdb.Close()

	err = pgdb.CreateTable(&Link{}, &orm.CreateTableOptions{})
	if err != nil {
		test.Failf("postgres-operator", "failed to create test table")
		return fmt.Errorf("failed to create table links: %v", err)
	}

	links := []interface{}{
		&Link{URL: "http://flanksource.com"},
		&Link{URL: "http://kubernetes.io"},
	}
	return pgdb.Insert(links...)
}

func testFixturesArePresent(p *platform.Platform, clusterName string, timeout time.Duration, test *console.TestResults) error {
	pgdb, err := p.OpenDB(Namespace, clusterName, "postgres")
	if err != nil {
		test.Failf("postgres-operator", "failed to connect to postgres")
		return err
	}
	defer pgdb.Close()

	deadline := time.Now().Add(timeout)

	for {
		var links []Link
		err := pgdb.Model(&links).Select()
		if err != nil {
			test.Errorf("failed to list links: %v", err)
		} else if len(links) != 2 {
			test.Errorf("expected 2 links got %d", len(links))
		} else {
			return nil
		}
		time.Sleep(5 * time.Second)
		if time.Now().After(deadline) {
			test.Failf("postgres-operator", "deadline exceeded waiting for links to be present in cloned database")
			return fmt.Errorf("could not find any links in database postgres, deadline exceeded")
		}
	}
}

func listObjects(client *minio.Client, bucket, path string) []minio.ObjectInfo {
	var objects []minio.ObjectInfo
	doneCh := make(chan struct{})
	defer close(doneCh)
	for obj := range client.ListObjectsV2(bucket, path, true, doneCh) {
		objects = append(objects, obj)
	}
	return objects
}

// func waitForWalBackup(p *platform.Platform, clusterName string, timeout time.Duration, timestamp time.Time, test *console.TestResults) error {
// 	client, err := p.GetS3Client()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get aws client")
// 	}

// 	bucket := getBackupBucket(p)
// 	deadline := time.Now().Add(timeout)
// 	baseBackupPath := fmt.Sprintf("%s/wal/basebackups_005", clusterName)
// 	walPath := fmt.Sprintf("%s/wal/wal_005", clusterName)

// 	for {
// 		if time.Now().After(deadline) {
// 			break
// 		}

// 		if len(listObjects(client, bucket, baseBackupPath)) == 0 {
// 			test.Infof("Did not find any base backups in bucket %s, retrying in 5 seconds", bucket)
// 			time.Sleep(5 * time.Second)
// 			continue
// 		}

// 		for _, wal := range listObjects(client, bucket, walPath) {
// 			if wal.LastModified.After(timestamp) {
// 				return nil
// 			}
// 		}
// 	}
// 	return fmt.Errorf("did not find base backup and/or wals in bucket %s, deadline exceeded", bucket)
// }

// func checkReplicaLag(p *platform.Platform, clusters ...string) error {
// 	for _, cluster := range clusters {
// 		patroniClient, err := GetPatroniClient(p, Namespace, cluster)
// 		if err != nil {
// 			return errors.Errorf("Failed to get patroni client to cluster %s", cluster)
// 		}
// 		response, err := patroniClient.Get("http://patroni/cluster")
// 		if err != nil {
// 			return errors.Errorf("Failed to get /cluster endpoint for cluster %s: %v", cluster, err)
// 		}
// 		defer response.Body.Close() // nolint: errcheck
// 		clusterResponse := &ClusterResponse{}
// 		err = json.NewDecoder(response.Body).Decode(&clusterResponse)
// 		if err != nil {
// 			return errors.Errorf("Failed to read response body for cluster %s: %v", cluster, err)
// 		}

// 		for _, m := range clusterResponse.Members {
// 			if m.State != "running" {
// 				return errors.Errorf("Expected state for cluster=%s node=%s to be 'running', got %s", cluster, m.Name, m.State)
// 			} else if m.Role == "replica" {
// 				iLag, ok := m.Lag.(int)
// 				if ok && iLag > 0 {
// 					return errors.Errorf("Expected replication lag for cluster=%s replica=%s to be 0, got %d", cluster, m.Name, m.Lag)
// 				} else if !ok {
// 					sLag, ok := m.Lag.(string)
// 					if ok && sLag != "" {
// 						return errors.Errorf("Expected replication lag for cluster=%s replica=%s to be 0, got %s", cluster, m.Name, m.Lag)
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func removeE2ECluster(p *platform.Platform, config pgapi.ClusterConfig, test *console.TestResults) {
	clusterName := "postgres-" + config.Name
	db := pgapi.NewPostgresql(clusterName)

	pgClient, _, _, err := p.GetDynamicClientFor(Namespace, db)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}

	_, err = pgClient.Get(context.TODO(), clusterName, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return
	}

	if err := pgClient.Delete(context.TODO(), clusterName, metav1.DeleteOptions{}); err != nil {
		test.Warnf("Failed to delete resource %s/%s/%s in namespace %s", db.APIVersion, db.Kind, db.Name, config.Namespace)
		return
	}
	test.Infof("Deleted pg cluster: %s", clusterName)
}
