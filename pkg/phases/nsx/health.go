package nsx

import (
	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.NSX == nil || p.NSX.Disabled {
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "nsx-system", test)
}
