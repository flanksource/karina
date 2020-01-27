package pgo

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.PGO == nil || p.PGO.Disabled {
		test.Skipf("PGO", "PGO not configured")
		return
	}
	k8s.TestNamespace(client, "pgo", test)
}
