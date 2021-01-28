package istiooperator

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.IstioOperator.IsDisabled() {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(Namespace, "Could not connect to Platform client: %v", err)
		return
	}
	p.WaitForNamespace(Namespace, 60*time.Second)
	kommons.TestNamespace(client, Namespace, test)
}
