package argocdoperator

import (
	"fmt"
	"time"

	"github.com/flanksource/commons/console"
	argo "github.com/flanksource/karina/pkg/api/argocd"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return
	}
	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "argocd"
	clusterName := "argocd"

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
