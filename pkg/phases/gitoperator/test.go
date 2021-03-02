package gitoperator

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.GitOperator.IsDisabled() {
		return
	}

	client, _ := p.GetClientset()
	kommons.TestDeploy(client, Namespace, "git-operator-controller-manager", test)
}
