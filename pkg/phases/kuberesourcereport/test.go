package kuberesourcereport

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func TestKubeResourceReport(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.KubeResourceReport == nil || p.KubeResourceReport.Disabled {
		test.Skipf("kube-resource-report", "kube-resource-report is not configured")
		return
	}
	k8s.TestNamespace(client, Namespace, test)
	k8s.TestDeploy(client, Namespace, "kube-resource-report", test)
}
