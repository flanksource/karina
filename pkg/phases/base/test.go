package base

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
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

	kommons.TestNamespace(client, "kube-system", test)
	kommons.TestNamespace(client, "local-path-storage", test)
	kommons.TestNamespace(client, "cert-manager", test)

	if platform.Nginx == nil || !platform.Nginx.Disabled {
		platform.WaitForNamespace("ingress-nginx", 180*time.Second)
		kommons.TestNamespace(client, "ingress-nginx", test)
	}
}
