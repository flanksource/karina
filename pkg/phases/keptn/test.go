package keptn

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Keptn.IsDisabled() {
		return
	}
	testName := "Keptn"
	expectedKeptnDeployments := []string{
		"api-gateway-nginx",
		"api-service",
		"bridge",
		"configuration-service",
		"eventbroker-go",
		"lighthouse-service",
		"mongodb-datastore",
		"remediation-service",
		"shipyard-service",
	}
	for _, deployName := range expectedKeptnDeployments {
		err := p.WaitForDeployment(Namespace, deployName, 1*time.Minute)
		if err != nil {
			test.Failf(testName, "Keptn component %s (Deployment) is not healthy: %v", deployName, err)
			return
		}
	}
	expectedKeptnStatefulsets := []string{
		"keptn-nats-cluster",
		"mongodb-keptn",
	}
	for _, statefulsetName := range expectedKeptnStatefulsets {
		err := p.WaitForStatefulSet(Namespace, statefulsetName, 1*time.Minute)
		if err != nil {
			test.Failf(testName, "Keptn component %s (StatefulSet) is not healthy: %v", statefulsetName, err)
			return
		}
	}
	test.Passf(testName, "Keptn is healthy")
}
