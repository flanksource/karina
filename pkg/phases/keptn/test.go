package keptn

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Keptn.IsDisabled() {
		test.Skipf("keptn", "Keptn is disabled")
		return
	}
	expectedKeptnDeployments := []string{
		"api-gateway-nginx",
		"api-service",
		"bridge",
		"configuration-service",
		"eventbroker-go",
		"lighthouse-service",
		"mongodb-datastore",
		"mongodb-keptn",
		"remediation-service",
		"shipyard-service",
	}
	for _, deployName := range expectedKeptnDeployments {
		err := p.WaitForDeployment(Namespace, deployName, 1*time.Minute)
		if err != nil {
			test.Failf("Keptn", "Keptn component %s is not healthy: %v", deployName, err)
			return
		}
	}
	test.Passf("Keptn", "Keptn is healthy")
}
