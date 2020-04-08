package postgresoperator

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-pg/pg/v9/orm"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	pg "github.com/go-pg/pg/v9"
	pgapi "github.com/moshloop/platform-cli/pkg/api/postgres"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Link struct {
	ID  int64
	URL string
}

func Test(p *platform.Platform, test *console.TestResults) {
	if p.PostgresOperator == nil || p.PostgresOperator.Disabled {
		test.Skipf("postgres-operator", "Postgres operator is disabled")
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, Namespace, test)
	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "postgres-operator-e2e"
	if p.PostgresOperator == nil || p.PostgresOperator.Disabled {
		test.Skipf(testName, "Postgres operator is disabled")
		return
	}
	client, _ := p.GetClientset()
	cluster1 := pgapi.NewClusterConfig(utils.RandomString(6), "test", "e2e_db")
	cluster1.BackupSchedule = "*/1 * * * *"
	cluster1Name := "postgres-" + cluster1.Name
	db, err := p.GetOrCreateDB(cluster1)
	defer removeE2ECluster(p, cluster1) //failsafe removal of cluster
	if err != nil {
		test.Failf(testName, "Error creating db %s: %v", cluster1.Name, err)
		return
	}
	test.Passf(testName, "Cluster %s deployed", cluster1Name)

	pf1, port1, err := exposeService(p, cluster1Name)
	if err != nil {
		test.Failf(testName, "Failed to expose service %s: %v", cluster1Name, err)
		return
	}
	defer pf1.Process.Kill() // nolint: errcheck

	if err := insertTestFixtures(db, port1); err != nil {
		test.Failf(testName, "Failed to insert fixtures into database %s: %v", cluster1.Name, err)
		time.Sleep(320 * time.Second)
		return
	}
	timestamp := time.Now().Add(5 * time.Second)

	if err := waitForWalBackup(p, cluster1Name, 5*time.Minute, timestamp); err != nil {
		test.Failf(testName, "Failed to find any wal backups for database %s: %v", cluster1.Name, err)
		return
	}

	removeE2ECluster(p, cluster1)

	cluster2 := pgapi.NewClusterConfig(cluster1.Name+"-clone", "test")
	cluster2.Clone = &pgapi.CloneConfig{
		ClusterName: cluster1Name,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05 UTC"),
	}
	cluster2Name := "postgres-" + cluster2.Name
	db2, err := p.GetOrCreateDB(cluster2)
	defer removeE2ECluster(p, cluster2)
	if err != nil {
		test.Failf(testName, "Error creating db %s: %v", cluster2.Name, err)
		return
	}
	test.Passf(testName, "Cluster %s deployed user", cluster1.Name)

	pf2, port2, err := exposeService(p, cluster2Name)
	if err != nil {
		test.Failf(testName, "Failed to expose service %s: %v", cluster2Name, err)
		return
	}
	defer pf2.Process.Kill() // nolint: errcheck

	if err := testFixturesArePresent(db2, port2, 5*time.Minute); err != nil {
		test.Failf(testName, "Failed to find test fixtures data in clone database %s: %v", cluster2Name, err)
		return
	}
	test.Passf(testName, "Cloned cluster %s successfully created", cluster2Name)
}

func insertTestFixtures(db *types.DB, port int) error {
	pgdb := pg.Connect(&pg.Options{
		User:     db.Username,
		Password: db.Password,
		Addr:     fmt.Sprintf("localhost:%d", port),
		Database: "e2e_db",
	})
	defer pgdb.Close()

	err := pgdb.CreateTable(&Link{}, &orm.CreateTableOptions{})
	if err != nil {
		return fmt.Errorf("failed to create table links: %v", err)
	}

	links := []interface{}{
		&Link{URL: "http://flanksource.com"},
		&Link{URL: "http://kubernetes.io"},
	}
	err = pgdb.Insert(links...)
	return fmt.Errorf("failed to insert links: %v", err)
}

func testFixturesArePresent(db *types.DB, port int, timeout time.Duration) error {
	pgdb := pg.Connect(&pg.Options{
		User:     db.Username,
		Password: db.Password,
		Addr:     fmt.Sprintf("localhost:%d", port),
		Database: "e2e_db",
	})
	defer pgdb.Close()

	deadline := time.Now().Add(timeout)

	for {
		var links []Link
		err := pgdb.Model(&links).Select()
		if err != nil {
			log.Errorf("failed to list links: %v", err)
		} else if len(links) != 2 {
			log.Errorf("expected 2 links got %d", len(links))
		} else {
			return nil
		}
		time.Sleep(5 * time.Second)
		if time.Now().After(deadline) {
			return fmt.Errorf("could not find any links in database e2e_db, deadline exceeded")
		}
	}
}

func waitForWalBackup(p *platform.Platform, clusterName string, timeout time.Duration, timestamp time.Time) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cfg := aws.NewConfig().
		WithRegion(p.S3.Region).
		WithEndpoint(p.S3.GetExternalEndpoint()).
		WithCredentials(
			credentials.NewStaticCredentials(p.S3.AccessKey, p.S3.SecretKey, ""),
		).
		WithHTTPClient(&http.Client{Transport: tr})
	ssn, err := session.NewSession(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create S3 session")
	}
	client := s3.New(ssn)
	client.Config.S3ForcePathStyle = aws.Bool(p.S3.UsePathStyle)

	bucket := p.PostgresOperatorBackupBucket()
	deadline := time.Now().Add(timeout)
	paths := []string{
		fmt.Sprintf("%s/wal/basebackups_005", clusterName),
		fmt.Sprintf("%s/wal/wal_005", clusterName),
	}

	for {
		foundAll := true
		walSegments := []*s3.Object{}
		for i, path := range paths {
			req := &s3.ListObjectsInput{
				Bucket:  aws.String(bucket),
				Marker:  aws.String(path),
				MaxKeys: aws.Int64(500),
			}
			resp, err := client.ListObjects(req)
			if err != nil {
				return fmt.Errorf("failed to list objects in bucket %s: %v", bucket, err)
			}
			if len(resp.Contents) == 0 || !strings.HasPrefix(aws.StringValue(resp.Contents[0].Key), path) {
				log.Infof("Did not find any object in bucket %s, retrying in 5 seconds", bucket)
				foundAll = false
				continue
			} else {
				log.Debugf("Found key %s for prefix %s, backups found", aws.StringValue(resp.Contents[0].Key), path)
				if i == 1 {
					// save wal segments objects to check for latest timestamp
					walSegments = resp.Contents
				}
			}
		}
		if foundAll {
			for _, o := range walSegments {
				// Check if latest backup file timestamp is after we inserted the test data
				if strings.HasPrefix(aws.StringValue(o.Key), paths[1]) && aws.TimeValue(o.LastModified).After(timestamp) {
					return nil
				}
			}
		}
		time.Sleep(5 * time.Second)
		if time.Now().After(deadline) {
			return fmt.Errorf("could not find any backups in bucket %s, deadline exceeded", bucket)
		}
	}
}

func exposeService(p *platform.Platform, clusterName string) (*exec.Cmd, int, error) {
	client, _ := p.GetClientset()
	opts := metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster-name=%s,spilo-role=master", clusterName)}
	pods, err := client.CoreV1().Pods(Namespace).List(opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get master pod for cluster %s: %v", clusterName, err)
	}

	if len(pods.Items) != 1 {
		return nil, 0, fmt.Errorf("expected 1 pod for spilo-role=master got %d", len(pods.Items))
	}

	randomPort := 36000 + rand.Intn(1000)

	portForwardCmd := exec.Command("./.bin/kubectl", "--namespace", "postgres-operator", "port-forward", pods.Items[0].Name, fmt.Sprintf("%d:5432", randomPort))
	if err := portForwardCmd.Start(); err != nil {
		return nil, 0, fmt.Errorf("failed to start portforward cmd: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)

	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", randomPort), 3*time.Second)
		if conn != nil {
			defer conn.Close()
		}

		if err, ok := err.(*net.OpError); ok && err.Timeout() {
			log.Errorf("Timeout error connecting to port %d: %s, retrying in 1 second", randomPort, err)
		} else if err != nil {
			// Log or report the error here
			log.Warnf("Error connecting to port %d: %s, retrying in 1 second", randomPort, err)
		} else {
			return portForwardCmd, randomPort, nil
		}
		time.Sleep(1 * time.Second)
		if time.Now().After(deadline) {
			portForwardCmd.Process.Kill() // nolint: errcheck
			return nil, 0, fmt.Errorf("timed out connecting to port %d", randomPort)
		}
	}
}

func removeE2ECluster(p *platform.Platform, config pgapi.ClusterConfig) {
	clusterName := "postgres-" + config.Name
	db := pgapi.NewPostgresql(clusterName)

	pgClient, _, _, err := p.GetDynamicClientFor(Namespace, db)
	if err != nil {
		log.Errorf("Failed to get dynamic client: %v", err)
		return
	}

	existing, err := pgClient.Get(clusterName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return
	}

	if err := pgClient.Delete(clusterName, nil); err != nil {
		log.Warnf("Failed to delete resource %s/%s/%s in namespace %s", db.APIVersion, db.Kind, db.Name, config.Namespace)
		return
	}
}
