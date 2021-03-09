package mongodboperator

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/constants"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
)

const (
	testNamespace = "test-mongodb-operator"
)

func Test(p *platform.Platform, test *console.TestResults) {
	testName := "mongodb-operator"
	if p.MongodbOperator.IsDisabled() {
		test.Skipf(testName, "MongoDB Operator is disabled")
		return
	}

	expectedMongoDBOperatorDeployments := []string{
		"percona-server-mongodb-operator",
	}
	for _, deployName := range expectedMongoDBOperatorDeployments {
		err := p.WaitForDeployment(constants.PlatformSystem, deployName, 1*time.Minute)
		if err != nil {
			test.Failf(testName, "MongoDB Operator component %s (Deployment) is not healthy: %v", deployName, err)
			return
		}
	}
	test.Passf(testName, "MongoDB Operator is healthy")

	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "mongodb-operator-e2e"
	clusterName := "my-cluster-name"

	defer removeE2ETestResources(p, test)

	if err := p.CreateOrUpdateNamespace(testNamespace, nil, nil); err != nil {
		test.Failf(testName, "Failed to create test namespace %s", testNamespace)
		return
	}
	if err := p.ApplySpecs(testNamespace, "test/percona-server-mongodb.yaml"); err != nil {
		test.Failf(testName, "Error creating PerconaServerMongoDB object: %v", err)
		return
	}

	test.Infof("Checking MongoDB Cluster's health...")
	client, _ := p.GetClientByKind("PerconaServerMongoDB")
	timeout := 3 * time.Minute
	start := time.Now()
	for {
		time.Sleep(1 * time.Second)
		item, err := client.Namespace(testNamespace).Get(context.TODO(), clusterName, metav1.GetOptions{})

		if err != nil {
			test.Debugf("Unable to get PerconaServerMongoDB/%s: %v", clusterName, err)
			continue
		}

		status := item.Object["status"]
		if status == nil {
			// Continue waiting if the status field hasn't been populated yet
			continue
		}

		state := item.Object["status"].(map[string]interface{})["state"]
		if state == nil {
			// Continue waiting if the status.ready field hasn't been populated yet
			continue
		} else if state == "ready" {
			test.Passf(testName, "PerconaServerMongoDB/%s status.state has been updated to ready (by MongoDB Operator)", clusterName)
			return
		}

		if start.Add(timeout).Before(time.Now()) {
			test.Failf(testName, "PerconaServerMongoDB's status.state has been updated to ready (by MongoDB Operator). within allowed time. MongoDB Operator is not running?")
			return
		}
	}
}

func removeE2ETestResources(p *platform.Platform, test *console.TestResults) {
	if err := p.DeleteSpecs(testNamespace, "test/percona-server-mongodb.yaml"); err != nil {
		test.Warnf("Failed to cleanup MongoDB Operator test resources in namespace %s", testNamespace)
	}

	client, _ := p.GetClientset()
	err := client.CoreV1().Namespaces().Delete(context.TODO(), testNamespace, metav1.DeleteOptions{})
	if err != nil {
		test.Warnf("Failed to delete test namespace %s", testNamespace)
	}
	test.Infof("Finished cleanup MongoDB Operator test resources")
}
