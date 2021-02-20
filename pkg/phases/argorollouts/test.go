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
		test.Skipf("argorollouts", "Argo Rollouts is disabled")
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
		test.Failf(testName, "Error creating Rollout object: %v", err)
		return
	}

	test.Infof("Checking if Argo Rollouts is working...")
	client, _ := p.GetClientByKind("Rollout")
	timeout := 2 * time.Minute
	start := time.Now()
	for {
		time.Sleep(1 * time.Second)

		item, err := client.Namespace(TestNamespace).Get(context.TODO(), rolloutName, metav1.GetOptions{})

		if err != nil {
			test.Debugf("Unable to get Rollout/%s: %v", rolloutName, err)
			continue
		}

		status := item.Object["status"]
		if status == nil {
			// Continue waiting if the status field hasn't been populated yet
			continue
		}

		conditions := item.Object["status"].(map[string]interface{})["conditions"].([]interface{})
		if conditions == nil {
			// Continue waiting if the status.conditions field hasn't been populated yet
			continue
		}

		for _, raw := range conditions {
			condition := raw.(map[string]interface{})
			if condition["type"] == "Available" && condition["status"] == "True" {
				test.Passf(testName, "Rollout object's status has been updated by Argo Rollouts")
				return
			}
		}

		if start.Add(timeout).Before(time.Now()) {
			test.Failf(testName, "Rollout object's status wasn't updated by ArgoRollouts within allowed time. Argo Rollout is not running?")
			return
		}
	}
}

func removeE2ETestResources(p *platform.Platform, test *console.TestResults) {
	if err := p.DeleteSpecs(TestNamespace, "test/argo-rollouts.yaml"); err != nil {
		test.Warnf("Failed to cleanup Argo Rollouts test resources in namespace %s", TestNamespace)
	}

	client, _ := p.GetClientset()
	err := client.CoreV1().Namespaces().Delete(context.TODO(), TestNamespace, metav1.DeleteOptions{})
	if err != nil {
		test.Warnf("Failed to delete test namespace %s", TestNamespace)
	}
	test.Infof("Finished cleanup ArgoRollouts test resources")
}
