package argorollouts

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

const (
	TestNamespace = "test-argorollouts"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.ArgoRollouts.IsDisabled() {
		return
	}
	client, _ := p.GetClientset()
	kommons.TestDeploy(client, constants.PlatformSystem, "argo-rollouts", test)
	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "argorollouts-e2e"
	rolloutName := "rollouts-demo"

	defer removeE2ETestResources(p, test)
	if err := p.CreateOrUpdateNamespace(TestNamespace, nil, nil); err != nil {
		test.Failf(testName, "Failed to create test namespace %s", TestNamespace)
		return
	}
	if err := p.ApplySpecs(TestNamespace, "test/argo-rollouts.yaml"); err != nil {
		test.Failf(testName, "%v", err)
		return
	}

	if _, err := p.WaitForResource("Rollout", TestNamespace, rolloutName, 2*time.Minute); err != nil {
		test.Failf(testName, "argo rollout is not ready: %s", err)
		return
	}
	test.Passf(testName, "argo rollout is ready")
}

func removeE2ETestResources(p *platform.Platform, test *console.TestResults) {
	if p.PlatformConfig.Trace {
		return
	}
	if err := p.DeleteSpecs(TestNamespace, "test/argo-rollouts.yaml"); err != nil {
		test.Warnf("Failed to cleanup Argo Rollouts test resources in namespace %s", TestNamespace)
	}

	client, _ := p.GetClientset()
	err := client.CoreV1().Namespaces().Delete(context.TODO(), TestNamespace, metav1.DeleteOptions{})
	if err != nil {
		test.Warnf("Failed to delete test namespace %s", TestNamespace)
	}
}
