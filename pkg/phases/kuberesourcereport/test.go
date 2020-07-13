package kuberesourcereport

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func TestKubeResourceReport(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.Monitoring == nil {
		test.Skipf("monitoring", "monitoring is not configured")
		return
	}
	k8s.TestNamespace(client, "monitoring", test)
	k8s.TestDeploy(client, "monitoring", "kube-resource-report", test)
}
