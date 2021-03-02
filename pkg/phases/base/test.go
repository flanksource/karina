package base

import (
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
}
