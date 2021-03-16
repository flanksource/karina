package registrycreds

import (
	"context"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.RegistryCredentials == nil || p.RegistryCredentials.Disabled {
		return
	}
	client, _ := p.GetClientset()
	namespace := p.RegistryCredentials.Namespace
	// HACK: registry-creds creates secrets for namespaces in alphabetical order
	// we want to create our secret faster
	testNamespace := "0001-test-registry-creds"
	secretName := "awsecr-cred"

	kommons.TestNamespace(client, namespace, test)

	if !p.E2E {
		return
	}

	if err := p.CreateOrUpdateWorkloadNamespace(testNamespace, nil, nil); err != nil {
		test.Failf("registry-creds", "failed to create namespace %s", testNamespace)
		return
	}
	defer func() {
		_ = client.CoreV1().Namespaces().Delete(context.TODO(), testNamespace, metav1.DeleteOptions{})
	}()

	// wait for up to 4 minutes for registry-credentials to create the secrets
	// in the background
	_ = wait.PollImmediate(1*time.Second, 4*time.Minute, func() (bool, error) {
		p.Debugf("Checking for pull secret: %s", secretName)
		return p.HasSecret(testNamespace, secretName), nil
	})

	secret, err := client.CoreV1().Secrets(testNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		test.Failf("registry-creds", "failed to get secret %s in namespace %s: %v", secretName, testNamespace, err)
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
}
