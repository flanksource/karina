package kpack

import (
	"fmt"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/exec"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"strings"
	"time"
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
	// Wait for Builder status to be true -> wait for 5 mintues
	for i := 1; i < 5; i++ {
	output, ok := exec.SafeExec(fmt.Sprintf("kubectl get builder test-builder -n %s", Namespace))
	if !ok{
		test.Failf(testName, "Failed to check builder status")
		return
	}
	contains := strings.Contains(output, "True")
	if contains {
		continue
	}
	time.Sleep(5*time.Minute)
	}

	// Wait for Image status to be True -> wait for 10 minutes
	for i := 1; i < 5; i++ {
		output, ok := exec.SafeExec(fmt.Sprintf("kubectl get image test-image -n %s", Namespace))
		if !ok {
			test.Failf(testName, "Failed to check image status")
			return
		}
		contains := strings.Contains(output, "True")
		if contains {
			test.Passf(testName, "image is ready")
			return
		}
	}
}