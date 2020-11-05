package kuberesourcereport

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func TestKubeResourceReport(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.KubeResourceReport == nil || p.KubeResourceReport.Disabled {
		test.Skipf("kube-resource-report", "kube-resource-report is not configured")
		return
	}
	kommons.TestNamespace(client, Namespace, test)
	kommons.TestDeploy(client, Namespace, "kube-resource-report", test)
}
