package canary

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

const testName = "canary-checker"

func TestCanary(p *platform.Platform, test *console.TestResults) {
	if p.CanaryChecker == nil || p.CanaryChecker.Disabled {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(testName, "couldn't get clientset: %v", err)
		return
	}

	k8s.TestDeploy(client, constants.PlatformSystem, "canary-checker", test)
}
