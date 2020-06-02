package nsx

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.NSX == nil || p.NSX.Disabled {
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "nsx-system", test)
}
