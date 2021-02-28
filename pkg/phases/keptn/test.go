package keptn

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Keptn.IsDisabled() {
		test.Skipf("keptn", "Keptn is disabled")
		return
	}
	client, _ := p.GetClientset()
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
		kommons.TestDeploy(client, constants.PlatformSystem, deployName, test)
	}
	test.Passf("Keptn", "Keptn is healthy")
}
