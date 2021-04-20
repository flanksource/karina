package kpack

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Kpack.Disabled || &p.Kpack.ImageVersions == nil {
		return
	}

	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	if p.E2E{
		TestE2EKpack(p, test)
	}
}

func TestE2EKpack(p *platform.Platform, test *console.TestResults) {
	testName := "kpack-e2e-test"
	if err := p.ApplySpecs(Namespace, "kpack-e2e.yaml"); err != nil{
		test.Failf(testName, "Failed to apply %v", err)
		return
	}
}