package contour

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.Contour == nil || p.Contour.Disabled {
		test.Skipf("Contour", "Contour not configured")
		return
	}
	k8s.TestNamespace(client, "contour", test)
}
