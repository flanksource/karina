package canary

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

const testName = "canary-checker"

func TestCanary(p *platform.Platform, test *console.TestResults) {
	if p.CanaryChecker.IsDisabled() {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(testName, "couldn't get clientset: %v", err)
		return
	}

	kommons.TestStatefulSet(client, constants.PlatformSystem, "canary-checker", test)
}
