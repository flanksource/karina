package flux

import (
	"context"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Import kubernetes oidc auth plugin
)

func Test(p *platform.Platform, test *console.TestResults) {
	if !p.E2E {
		return
	}

	testName := "gitops"
	namespace := "gitops-e2e-test"
	client, _ := p.GetClientset()
	fixture := types.GitOps{
		Name:                "karina",
		Namespace:           namespace,
		HelmOperatorVersion: "1.2.0",
		GitURL:              "https://github.com/flanksource/gitops-test.git",
		SyncInterval:        "5s",
		GitPollInterval:     "5s",
	}

	if err := p.CreateOrUpdateNamespace(namespace, nil, nil); err != nil {
		test.Failf(testName, "failed to create namespace: %v", err)
	}
	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	if err := p.Apply(namespace, NewFluxDeployment(&fixture)...); err != nil {
		test.Failf(testName, "failed to deploy gitops: %v", err)
		return
	}

	if err := p.WaitForDeployment(namespace, "flux-karina", 30*time.Second); err != nil {
		test.Failf(testName, "failed to deploy flux: %v", err)
	}

	err := p.WaitForDeployment(namespace, "nginx", 120*time.Second)
	if err != nil {
		test.Failf(testName, "Deployment 'nginx' not ready in namespace %s: %v", namespace, err)
	}
	err = p.WaitForDeployment(namespace, "gitops-e2e-test-podinfo", 120*time.Second)
	if err != nil {
		test.Failf(testName, "Deployment 'gitops-e2e-test-podinfo' not ready in namespace %s: %v", namespace, err)
	}

	kommons.TestNamespace(client, namespace, test)

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=nginx"})
	if err != nil {
		test.Failf("gitops", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 2 {
		test.Failf("gitops", "Expected 2 nginx pods in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("gitops", "Pods for deployment nginx created successfully")
	}

	pods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=podinfo"})
	if err != nil {
		test.Failf("helm-operator", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 1 {
		test.Failf("helm-operator", "Expected 1 podinfo in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("helm-operator", "Pods for podinfo helm chart created successfully")
	}

	if _, err = client.CoreV1().Services(namespace).Get(context.TODO(), "nginx", metav1.GetOptions{}); err != nil {
		test.Failf("gitops", "Failed to get service nginx in namespace %s: %v", namespace, err)
	} else {
		test.Passf("gitops", "Service nginx created successfully")
	}
}
