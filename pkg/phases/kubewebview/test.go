package kubewebview

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func TestKubeWebView(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.KubeWebView == nil || p.KubeWebView.Disabled {
		test.Skipf("kube-web-view", "kube-web-view is not configured")
		return
	}
	kommons.TestNamespace(client, Namespace, test)
	kommons.TestDeploy(client, Namespace, "kube-web-view", test)
}
