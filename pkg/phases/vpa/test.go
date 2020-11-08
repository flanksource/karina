package vpa

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.VPA.IsDisabled() {
		return
	}
	client, _ := p.GetClientset()

	for _, deployment := range []string{"vpa-admission-controller", "vpa-recommender", "vpa-updater"} {
		kommons.TestDeploy(client, Namespace, deployment, test)
	}
}
