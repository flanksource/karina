package certmanager

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
}
