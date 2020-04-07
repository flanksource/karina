package consul

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	Namespace = "vault"
)

func TestE2E(p *platform.Platform, test *console.TestResults) {
	if p.Vault == nil || p.Vault.Disabled {
		test.Skipf("consul", "Consul is disabled")
		return
	}

	cs := NewBackupRestore(p, "consul-server", "vault")

	key1 := "e2e-test-" + utils.RandomString(6)
	value1 := utils.RandomString(10)
	_, _, err := p.ExecutePodf("vault", "consul-server-0", "consul", "consul", "kv", "put", key1, value1)

	if err != nil {
		test.Failf("consul", "failed to set key %s to value %s: %v", key1, value1, err)
		return
	}

	test.Passf("consul", "set key %s=%s", key1, value1)

	if getConsulValue(p, "vault", "consul-server-0", "consul", key1) != value1 {
		test.Failf("consul", "expected key %s to equal %s", key1, value1)
		return
	}

	test.Passf("consul", "verified key %s=%s", key1, value1)

	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")

	if err := cs.Backup(); err != nil {
		test.Failf("consul", "failed to backup consul")
		return
	}

	test.Passf("consul", "backed up consul-server-0")

	value2 := utils.RandomString(10)
	_, _, err = p.ExecutePodf("vault", "consul-server-0", "consul", "consul", "kv", "put", key1, value2)

	if err != nil {
		test.Failf("consul", "failed to set key %s to value %s: %v", key1, value2, err)
		return
	}

	test.Passf("consul", "set key %s=%s", key1, value2)

	snapshotFilename, err := getSnapshotFilename(p, "vault", "consul-server", timestamp)
	if err != nil {
		test.Failf("consul", "failed to get snapshot filename: %v", err)
		return
	}

	test.Passf("consul", "found snapshot %s", snapshotFilename)

	if getConsulValue(p, "vault", "consul-server-0", "consul", key1) != value2 {
		test.Failf("consul", "expected key %s to equal %s", key1, value2)
		return
	}

	test.Passf("consul", "verified key %s=%s", key1, value2)

	if err := cs.Restore(snapshotFilename); err != nil {
		test.Failf("consul", "failed to restore consul from backup: %v", err)
		return
	}

	test.Passf("consul", "restored consul-server-0")

	if getConsulValue(p, "vault", "consul-server-0", "consul", key1) != value1 {
		test.Failf("consul", "expected key %s to equal %s", key1, value1)
		return
	}

	test.Passf("consul", "verified key %s=%s", key1, value1)

	test.Passf("consul", "Successfully ran backup/restore e2e tests")
}

func getConsulValue(p *platform.Platform, namespace, pod, container, key string) string {
	stdout, _, err := p.ExecutePodf(namespace, pod, container, "consul", "kv", "get", key)
	if err != nil {
		log.Errorf("failed to get key %s: %v", key, err)
		return ""
	}
	value := strings.TrimSuffix(stdout, "\n")
	return value
}

func getSnapshotFilename(p *platform.Platform, namespace, pod, timestamp string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cfg := aws.NewConfig().
		WithRegion(p.S3.Region).
		WithEndpoint(p.S3.ExternalEndpoint).
		WithCredentials(
			credentials.NewStaticCredentials(p.S3.AccessKey, p.S3.SecretKey, ""),
		).
		WithHTTPClient(&http.Client{Transport: tr})
	ssn, err := session.NewSession(cfg)
	if err != nil {
		return "", errors.Wrap(err, "failed to create S3 session")
	}
	client := s3.New(ssn)
	client.Config.S3ForcePathStyle = aws.Bool(p.S3.UsePathStyle)

	path := fmt.Sprintf("consul/backups/%s/%s/%s", namespace, pod, timestamp)
	directoryPath := fmt.Sprintf("consul/backups/%s/%s", namespace, pod)
	bucket := p.Vault.Consul.Bucket

	req := &s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		Marker:  aws.String(path),
		MaxKeys: aws.Int64(500),
	}

	resp, err := client.ListObjects(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to list snapshots in s3")
	}

	if len(resp.Contents) == 0 {
		return "", errors.Errorf("expected at least one snapshot")
	} else if !strings.HasPrefix(aws.StringValue(resp.Contents[0].Key), directoryPath) {
		return "", errors.Errorf("expected at least one snapshot in directory  %s, first element returned is %s", directoryPath, aws.StringValue(resp.Contents[0].Key))
	} else {
		snapshot := fmt.Sprintf("s3://%s/%s", bucket, aws.StringValue(resp.Contents[0].Key))
		return snapshot, nil
	}
}
