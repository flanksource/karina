package base

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	client, err := platform.GetClientset()
	if err != nil {
		test.Errorf("Base tests failed to get clientset: %v", err)
		return
	}
	if client == nil {
		test.Errorf("Base tests failed to get clientset: nil clientset ")
		return
	}

	k8s.TestNamespace(client, "kube-system", test)
	k8s.TestNamespace(client, "local-path-storage", test)
	k8s.TestNamespace(client, "cert-manager", test)

	if platform.Nginx == nil || !platform.Nginx.Disabled {
		k8s.TestNamespace(client, "ingress-nginx", test)
	}

}
