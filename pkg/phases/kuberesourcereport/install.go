package kuberesourcereport

import (
	"fmt"

	"time"

	"github.com/flanksource/karina/pkg/types"

	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = constants.PlatformSystem
	Group     = "system:reporting"
	User      = "kube-resource-report"
)

func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil {
		p.KubeResourceReport = &types.KubeResourceReport{} // this sets p.KubeResourceReport.Disabled to false
		p.KubeResourceReport.Disabled = true
	}
	if p.KubeResourceReport.Disabled {
		// remove the secret containing access information to external clusters
		cs, err := p.GetClientset()
		if err != nil {
			return fmt.Errorf("failed to get clientset for cluster: %v", err)
		}
		err = cs.CoreV1().Secrets(Namespace).Delete("kube-resource-report-clusters", &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove external cluster access secret: %v", err)
		}
		return p.DeleteSpecs(Namespace, "kube-resource-report.yaml")
	}

	// make sure the namespace exists
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	ca, err := ca.ReadCA(p.CA)
	if err != nil {
		return fmt.Errorf("unable to get root CA %v", err)
	}
	// kube-resource-view can't use the service account to access it's own cluster
	// so we add user/cert access via the default internal API endpoint
	p.KubeResourceReport.ExternalClusters.AddSelf(p.Name)
	// create a secret containing a kubeconfig file that allows access to
	// this cluster via user/cert as well as the given external clusters
	kubeConfig, err := k8s.CreateMultiKubeConfig(ca, p.KubeResourceReport.ExternalClusters, Group, User, 24*7*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to generate kubeconfig for multi-cluster access: %v", err)
	}
	if p.PlatformConfig.Trace {
		p.Infof("kubeconfig file is:\n%v", string(kubeConfig))
	}
	err = p.CreateOrUpdateSecret("kube-resource-report-clusters", Namespace, map[string][]byte{
		"config": kubeConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to generate kubeconfig secret for multi-cluster access: %v", err)
	}

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
