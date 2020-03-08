package configmapReloader

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.ConfigMapReloader == nil || p.ConfigMapReloader.Disabled {
		test.Skipf("configmap-reloader", "configmap-reloader not configured")
		return
	}
	k8s.TestNamespace(client, "configmap-reloader", test)
}