package eck

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.ECK == nil || p.ECK.Disabled {
		test.Skipf("ECK", "ECK not configured")
		return
	}
	k8s.TestNamespace(client, Namespace, test)
}
