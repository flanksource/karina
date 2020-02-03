package fluentdOperator

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.FluentdOperator == nil || p.FluentdOperator.Disabled {
		test.Skipf("FluentdOperator", "FluentdOperator not configured")
		return
	}
	k8s.TestNamespace(client, "kube-fluentd-operator", test)
}
