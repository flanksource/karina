package argocdoperator

import (
	"fmt"
	"time"

	"github.com/flanksource/commons/console"
	argo "github.com/flanksource/karina/pkg/api/argocd"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewArgoCDClusterConfig(name string) *argo.ArgoCD {
	return &argo.ArgoCD{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ArgoCD",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func Test(p *platform.Platform, test *console.TestResults) {
	if p.ArgoCDOperator.IsDisabled() {
		test.Skipf("argocd-operator", "ArgoCD Operator is disabled")
		return
	}
	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "argocd-operator-e2e"
	if p.ArgoCDOperator.IsDisabled() {
		test.Skipf(testName, "ArgoCD Operator is disabled")
		return
	}
	clusterName := "test-cluster"
	testCluster := NewArgoCDClusterConfig(clusterName)
	err := p.Apply(Namespace, testCluster)
	defer removeE2EArgoCluster(p, clusterName, test)
	if err != nil {
		test.Failf(testName, "Error creating ArgoCD Cluster %s: %v", clusterName, err)
		return
	}
	test.Passf(testName, "Cluster %s deployed", clusterName)

	test.Infof("Checking if ArgoCD cluster is healthy...")
	// List of expected deployment to be deployed by ArgoCD Operator for ArgoCD Cluster
	expectedDeployments := []string{
		fmt.Sprintf("%s-application-controller", clusterName),
		fmt.Sprintf("%s-dex-server", clusterName),
		fmt.Sprintf("%s-redis", clusterName),
		fmt.Sprintf("%s-repo-server", clusterName),
		fmt.Sprintf("%s-server", clusterName),
	}

	for _, deployName := range expectedDeployments {
		err := p.WaitForDeployment(Namespace, deployName, 1*time.Minute)
		if err != nil {
			test.Failf(testName, "ArgoCD Cluster component %s is not healthy: %v", deployName, err)
			return
		}
	}
	test.Passf(testName, "ArgoCD Cluster is healthy")
}

func removeE2EArgoCluster(p *platform.Platform, clusterName string, test *console.TestResults) {
	err := p.DeleteUnstructured(Namespace, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "ArgoCD",
			"metadata": map[string]interface{}{
				"name": clusterName,
			},
		},
	})
	if err != nil {
		test.Warnf("Failed to cleanup ArgoCD Test Cluster %s in namespace %s", clusterName, Namespace)
		return
	}
	test.Infof("Deleted ArgoCD cluster: %s", clusterName)
}
