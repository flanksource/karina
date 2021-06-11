package konfigmanager

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

const (
	testName = "konfigManager"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.CanaryChecker == nil || p.CanaryChecker.Disabled {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(testName, "couldn't get clientset: %v", err)
		return
	}

	kommons.TestDeploy(client, Namespace, "konfig-manager", test)
}
