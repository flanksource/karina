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
	if _, err := p.WaitForResource("PerconaServerMongoDB", testNamespace, clusterName, 3*time.Minute); err != nil {
		test.Failf(testName, "MongoDB Cluster %s is not ready within 3 minutes", clusterName)
		return
	}
	test.Passf(testName, "MongoDB Cluster %s is ready within 3 minutes", clusterName)
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
