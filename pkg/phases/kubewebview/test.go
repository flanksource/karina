package kubewebview

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func TestKubeWebView(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.KubeWebView == nil || p.KubeWebView.Disabled {
		test.Skipf("kube-web-view", "kube-web-view is not configured")
		return
	}
	k8s.TestNamespace(client, Namespace, test)
	k8s.TestDeploy(client, Namespace, "kube-web-view", test)
}
