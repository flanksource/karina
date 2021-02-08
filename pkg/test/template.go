package test

import (
	"context"
	"os"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	templateTestName        = "patch-template"
	templateTestNamespace   = "default"
	templateTestFixturePath = "test/fixtures/template-test.yaml"
	templateTestEnv         = "CONFIGURED_VALUE"
)

func TestTemplates(p *platform.Platform, test *console.TestResults) {
	client, err := p.GetClientset()
	if err != nil {
		test.Failf(templateTestName, "Couldn't get clientset: %v", err)
		return
	}

	if err := p.ApplyText(templateTestNamespace, files.SafeRead(templateTestFixturePath)); err != nil {
		test.Failf(templateTestName, "Failed to apply template test: %v", err)
		return
	}

	cmFile, err := client.CoreV1().ConfigMaps(templateTestNamespace).Get(context.TODO(), "template-test-file", metav1.GetOptions{})
	if err != nil {
		test.Failf(templateTestName, "couldn't get configmap templated from file: %v", err)
		return
	}
	if cmFile.Data["configuredValue"] != os.Getenv(templateTestEnv) {
		test.Failf(templateTestName, "patch file not templated. expected '%v', got '%v'", os.Getenv(templateTestEnv), cmFile.Data["configuredValue"])
		return
	}

	cmDirect, err := client.CoreV1().ConfigMaps(templateTestNamespace).Get(context.TODO(), "template-test-direct", metav1.GetOptions{})
	if err != nil {
		test.Failf(templateTestName, "couldn't get configmap templated from directly included patch: %v", err)
		return
	}
	if cmDirect.Data["configuredValue"] != os.Getenv(templateTestEnv) {
		test.Failf(templateTestName, "direct patch not templated. expected '%v', got '%v'", os.Getenv(templateTestEnv), cmFile.Data["configuredValue"])
		return
	}
	test.Passf(templateTestName, "patch and direct patch templated using '%v'", os.Getenv(templateTestEnv))
}
