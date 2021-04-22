package kpack

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Kpack.IsDisabled() {
		return
	}

	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	kommons.TestDeploy(client, Namespace, "kpack-controller", test) // Check if kpack-controller is running
	kommons.TestDeploy(client, Namespace, "kpack-webhook", test)    // Check if kpack-webhook is running
	if p.E2E {
		TestE2EKpack(p, test)
	}
}

func TestE2EKpack(p *platform.Platform, test *console.TestResults) {
	testName := "kpack-e2e-test"
	if err := p.ApplySpecs(Namespace, "kpack-e2e.yaml"); err != nil {
		test.Failf(testName, "Failed to apply %v", err)
		return
	}
	_, err := p.WaitForResource("Builder", Namespace, "test-builder", 5*time.Minute)
	if err != nil {
		test.Failf(testName, "Failed to check builder status")
		return
	}
	// Wait for Image status to be True -> wait for 10 minutes
	_, err = p.WaitForResource("Image", Namespace, "test-image", 10*time.Minute)
	if err != nil {
		test.Failf(testName, "Failed to check image status")
		return
	}
	test.Passf(testName, "image is ready")
}
