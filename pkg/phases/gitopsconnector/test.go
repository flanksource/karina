package gitopsconnector

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

const testName = "gitops-connector"

func Test(p *platform.Platform, test *console.TestResults) {
	if p.GitOpsConnector.IsDisabled() {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(testName, "couldn't get clientset: %v", err)
		return
	}

	kommons.TestDeploy(client, constants.PlatformSystem, "gitops-connector", test)
}
