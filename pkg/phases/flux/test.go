package flux

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
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

	err := p.WaitForDeployment(namespace, "nginx", 150*time.Second)
	if err != nil {
		test.Failf("gitops", "Deployment 'nginx' not ready in namespace %s: %v", namespace, err)
	}
	err = p.WaitForDeployment(namespace, "gitops-e2e-test-podinfo", 150*time.Second)
	if err != nil {
		test.Failf("gitops", "Deployment 'gitops-e2e-test-podinfo' not ready in namespace %s: %v", namespace, err)
	}

	k8s.TestNamespace(client, namespace, test)

	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=nginx"})
	if err != nil {
		test.Failf("gitops", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 2 {
		test.Failf("gitops", "Expected 2 nginx pods in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("gitops", "Pods for deployment nginx created successfully")
	}

	pods, err = client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=podinfo"})
	if err != nil {
		test.Failf("helm-operator", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 1 {
		test.Failf("helm-operator", "Expected 1 podinfo in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("helm-operator", "Pods for podinfo helm chart created successfully")
	}

	if _, err = client.CoreV1().Services(namespace).Get("nginx", metav1.GetOptions{}); err != nil {
		test.Failf("gitops", "Failed to get service nginx in namespace %s: %v", namespace, err)
	} else {
		test.Passf("gitops", "Service nginx created successfully")
	}
}
