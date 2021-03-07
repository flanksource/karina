package eck

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.ECK.IsDisabled() {
		return
	}
	kommons.TestNamespace(client, Namespace, test)
}
