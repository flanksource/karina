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
	if p.ArgocdOperator.IsDisabled() {
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
	if p.ArgocdOperator.IsDisabled() {
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

	// TODO: Temporarily only. We won't need this anymore once ArgoCD release a new version that incorporate this fix: https://github.com/argoproj-labs/argocd-operator/pull/224
	// Deploying necessary RBAC Objects for ArgoCD Cluster to fully working
	if err := p.ApplySpecs(Namespace, "argocd-rbac.yaml"); err != nil {
		test.Failf(testName, "Error creating RBAC Objects for ArgoCD Cluster: %v", err)
		return
	}

	test.Infof("Checking if ArgoCD cluster is healthy...")
	// List of expected deployment to be deployed by ArgoCD Operator for ArgoCD Cluster
	expectedDeploymentTypes := []string{
		"application-controller",
		"dex-server",
		"redis",
		"repo-server",
		"server",
	}

	for _, deployTypeName := range expectedDeploymentTypes {
		deploymentName := fmt.Sprintf("%s-%s", clusterName, deployTypeName)

		err := p.WaitForDeployment(Namespace, deploymentName, 1*time.Minute)
		if err != nil {
			test.Failf(testName, "ArgoCD Cluster component %s is not healthy: %v", deploymentName, err)
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
	}

	if err := p.DeleteSpecs("", "argocd-rbac.yaml"); err != nil {
		test.Warnf("Failed to delete ArgoCD Cluster RBAC Objects: %v", err)
	}

	test.Infof("Finished cleanup ArgoCD Cluster: %s", clusterName)
}
