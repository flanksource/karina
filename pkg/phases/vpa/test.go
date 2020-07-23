package vpa

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.VPA == nil || p.VPA.Version == "" {
		return
	}
	client, _ := p.GetClientset()

	for _, deployment := range []string{"vpa-admission-controller", "vpa-recommender", "vpa-updater"} {
		k8s.TestDeploy(client, Namespace, deployment, test)
	}
}
