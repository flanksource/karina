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
		"eventbroker-go",
		"lighthouse-service",
		"mongodb-datastore",
		"remediation-service",
		"shipyard-service",
	}

	// wait for up to 3 minutes for config service as it takes the longest
	if err := p.WaitForDeployment(Namespace, "configuration-service", 3*time.Minute); err != nil {
		test.Failf(testName, "ketpn/configuration-service is not healthy: %v", err)
	}

	// wait for (up to 3 minutes) + 30 seconds for everything else
	for _, deployName := range expectedKeptnDeployments {
		if err := p.WaitForDeployment(Namespace, deployName, 30*time.Second); err != nil {
			test.Failf(testName, "Keptn component %s (Deployment) is not healthy: %v", deployName, err)
		}
	}
	expectedKeptnStatefulsets := []string{
		"keptn-nats-cluster",
		"mongodb-keptn-rs0",
	}
	for _, statefulsetName := range expectedKeptnStatefulsets {
		if err := p.WaitForStatefulSet(Namespace, statefulsetName, 30*time.Second); err != nil {
			test.Failf(testName, "Keptn component %s (StatefulSet) is not healthy: %v", statefulsetName, err)
		}
	}
	test.Passf(testName, "Keptn is healthy")
}
