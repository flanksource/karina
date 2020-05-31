package fluentdoperator

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.FluentdOperator == nil || p.FluentdOperator.Disabled {
		test.Skipf("FluentdOperator", "FluentdOperator not configured")
		return
	}
	k8s.TestNamespace(client, "kube-fluentd-operator", test)
}
