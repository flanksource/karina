package registrycreds

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	namespace := p.RegistryCredentials.Namespace
	// HACK: registry-creds creates secrets for namespaces in alfabetically order
	// we want to create our secret faster
	testNamespace := "0001-test-registry-creds"
	secretName := "awsecr-cred"

	k8s.TestNamespace(client, namespace, test)

	if !p.E2E {
		return
	}

	if err := p.CreateOrUpdateNamespace(testNamespace, nil, nil); err != nil {
		test.Failf("registry-creds", "failed to create namespace %s", testNamespace)
		return
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		test.Failf("registry-creds", "failed to get secret %s in namespace %s", secretName, testNamespace)
		return
	}

	dockerOptions, ok := secret.Data[".dockerconfigjson"]
	if !ok {
		test.Failf("registry-creds", "failed to find .dockerconfigjson in secret %s in namespace %s", secretName, testNamespace)
		return
	}

	if string(dockerOptions) == "{}" {
		test.Failf("registry-creds", "expected secret %s in namespace %s to contain registry credentials", secretName, testNamespace)
		return
	}

	test.Passf("registry-creds", "secret %s in namespace %s has registry credentials", secretName, testNamespace)

	if err := p.ApplySpecs(testNamespace, "test-registry-creds.yaml"); err != nil {
		test.Failf("registry-creds", "failed to apply test-registry-creds.yaml")
		return
	}

	k8s.TestNamespace(client, testNamespace, test)
}
