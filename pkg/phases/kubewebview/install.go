package kubewebview

import (
	"fmt"
	"time"

	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const (
	Namespace = constants.PlatformSystem
	Group     = "system:reporting"
	User      = "kube-web-view"
)

func Install(p *platform.Platform) error {
	if p.KubeWebView == nil {
		p.KubeWebView = &types.KubeWebView{}
		return p.DeleteSpecs(Namespace, "kube-web-view.yaml")
	}
	if p.KubeWebView.Disabled {
		// if external clusters were configured we attempt to remove the configuration first
		if len(p.KubeWebView.ExternalClusters) > 0 {
			// remove the secret containing access information to external clusters
			cs, err := p.GetClientset()
			if err != nil || cs == nil {
				p.Warnf("failed to get clientset for cluster: %v", err)
			} else {
				err = cs.CoreV1().Secrets(Namespace).Delete("kube-web-view-clusters", &metav1.DeleteOptions{})
				if err != nil {
					p.Warnf("failed to remove external cluster access secret: %v", err)
				}
			}
		}
		return p.DeleteSpecs(Namespace, "kube-web-view.yaml")
	}

	// make sure the namespace exists
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	// if external clusters are specified we configure them to allow access
	if len(p.KubeWebView.ExternalClusters) > 0 {
		// we use our own root CA for ALL cluster accesses
		ca, err := ca.ReadCA(p.CA)
		if err != nil {
			return fmt.Errorf("unable to get root CA %v", err)
		}
		// kube-web-view can't use the service account to access it's own cluster
		// so we add user/cert access via the default internal API endpoint
		p.KubeWebView.ExternalClusters.AddSelf(p.Name)
		// create a secret containing a kubeconfig file that allows access to
		// this cluster via user/cert as well as the given external clusters
		kubeConfig, err := k8s.CreateMultiKubeConfig(ca, p.KubeWebView.ExternalClusters, Group, User, 24*7*time.Hour)
		if err != nil {
			return fmt.Errorf("failed to generate kubeconfig for multi-cluster access: %v", err)
		}
		if p.PlatformConfig.Trace {
			p.Infof("kubeconfig file is:\n%v", string(kubeConfig))
		}
		err = p.CreateOrUpdateSecret("kube-web-view-clusters", Namespace, map[string][]byte{
			"config": kubeConfig,
		})
		if err != nil {
			return fmt.Errorf("failed to generate kubeconfig secret for multi-cluster access: %v", err)
		}
	}
	return p.ApplySpecs(Namespace, "kube-web-view.yaml")
}
