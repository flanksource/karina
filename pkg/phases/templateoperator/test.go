package templateoperator

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.TemplateOperator.IsDisabled() {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(Namespace, "Could not connect to Platform client: %v", err)
		return
	}
	p.WaitForNamespace(Namespace, 60*time.Second)
	k8s.TestNamespace(client, Namespace, test)
}
