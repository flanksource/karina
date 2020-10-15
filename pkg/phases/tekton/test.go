package tekton

import (
	"github.com/flanksource/commons/console"

	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Tekton.Disabled || p.Tekton.Version == "" {
		return
	}

	client, _ := p.GetClientset()
	k8s.TestNamespace(client, Namespace, test)
}
