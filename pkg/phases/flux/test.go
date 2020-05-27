package flux

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Import kubernetes oidc auth plugin
)

func Test(p *platform.Platform, test *console.TestResults) {
	if !p.E2E {
		return
	}
	if len(p.GitOps) < 1 {
		test.Skipf("gitops", "No GitOps config specified - skipping.")
		return
	}
	namespace := "gitops-e2e-test"
	client, _ := p.GetClientset()

	k8s.TestNamespace(client, namespace, test)

	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=nginx"})
	if err != nil {
		test.Failf("gitops", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 2 {
		test.Failf("gitops", "Expected 2 nginx pods in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("gitops", "Pods for deployment nginx created successfully")
	}

	if _, err = client.CoreV1().Services(namespace).Get("nginx", metav1.GetOptions{}); err != nil {
		test.Failf("gitops", "Failed to get service nginx in namespace %s: %v", namespace, err)
	} else {
		test.Passf("gitops", "Service nginx created successfully")
	}
}
