package eck

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.PGO == nil || p.PGO.Disabled {
		test.Skipf("ECK", "ECK not configured")
		return
	}
	k8s.TestNamespace(client, "elastic-system", test)
}
