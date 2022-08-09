package flux

import (
	"context"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Import kubernetes oidc auth plugin
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Flux.Enabled {
		client, _ := p.GetClientset()
		kommons.TestNamespace(client, Namespace, test)
	}

	if !p.E2E {
		return
	}

	TestV2(p, test)
}

func TestV2(p *platform.Platform, test *console.TestResults) {
	if !p.Flux.Enabled {
		test.Skipf("FluxV2", "Flux v2 not enabled")
		return
	}

	testName := "flux-v2-e2e"
	namespace := "flux-v2-e2e"
	client, _ := p.GetClientset()

	gitRepository := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "source.toolkit.fluxcd.io/v1beta1",
			"kind":       "GitRepository",
			"metadata": map[string]interface{}{
				"name":      "flux-v2-e2e",
				"namespace": "flux-system",
			},
			"spec": map[string]interface{}{
				"interval": "20s",
				"reference": map[string]interface{}{
					"branch": "main",
				},
				"url": "https://github.com/flanksource/flux-v2-e2e",
			},
		},
	}

	if err := p.ApplyUnstructured(Namespace, &gitRepository); err != nil {
		test.Failf("FluxV2", "failed to apply GitRepository: %v", err)
		return
	}

	defer func() {
		p.DeleteByKind("GitRepository", "flux-system", "flux-v2-e2e") // nolint: errcheck
	}()

	kustomization := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1beta1",
			"kind":       "Kustomization",
			"metadata": map[string]interface{}{
				"name":      "flux-v2-e2e",
				"namespace": "flux-system",
			},
			"spec": map[string]interface{}{
				"interval": "20s",
				"path":     "test-1",
				"prune":    true,
				"sourceRef": map[string]interface{}{
					"kind": "GitRepository",
					"name": "flux-v2-e2e",
				},
				"validation": "client",
			},
		},
	}

	if err := p.Apply(Namespace, &kustomization); err != nil {
		test.Failf("FluxV2", "failed to apply Kustomization: %v", err)
		return
	}

	defer func() {
		p.DeleteByKind("Kustomization", "flux-system", "flux-v2-e2e") // nolint: errcheck
	}()

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	err := p.WaitForDeployment(namespace, "nginx", 120*time.Second)
	if err != nil {
		test.Failf(testName, "Deployment 'nginx' not ready in namespace %s: %v", namespace, err)
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=nginx"})
	if err != nil {
		test.Failf("gitops", "Failed to list pods in namespace %s: %v", namespace, err)
	} else if len(pods.Items) != 3 {
		test.Failf("gitops", "Expected 3 nginx pods in namespace %s got %d", namespace, len(pods.Items))
	} else {
		test.Passf("gitops", "Pods for deployment nginx created successfully")
	}

	if _, err = client.CoreV1().Services(namespace).Get(context.TODO(), "nginx", metav1.GetOptions{}); err != nil {
		test.Failf("gitops", "Failed to get service nginx in namespace %s: %v", namespace, err)
	} else {
		test.Passf("gitops", "Service nginx created successfully")
	}
}
