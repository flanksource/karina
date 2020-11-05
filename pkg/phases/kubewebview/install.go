package kubewebview

import (
	"fmt"
	"time"

	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const (
	Namespace     = constants.PlatformSystem
	Group         = "system:reporting"
	User          = "kube-web-view"
	ClusterConfig = "kube-web-view-clusters"
)

func Install(p *platform.Platform) error {
	if p.KubeWebView == nil {
		p.KubeWebView = &types.KubeWebView{} // this sets p.KubeWebView.Disabled to false
		p.KubeWebView.Disabled = true
	}
	if p.KubeWebView.Disabled {
		// remove the secret containing access information to external clusters
		cs, err := p.GetClientset()
		if err != nil {
			return err
		}
		if p.HasSecret(Namespace, ClusterConfig) {
			err = cs.CoreV1().Secrets(Namespace).Delete("kube-web-view-clusters", &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		return p.DeleteSpecs(Namespace, "kube-web-view.yaml")
	}

	// make sure the namespace exists
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

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
	kubeConfig, err := kommons.CreateMultiKubeConfig(ca, p.KubeWebView.ExternalClusters, Group, User, 2*356*24*time.Hour)
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

	return p.ApplySpecs(Namespace, "kube-web-view.yaml")
}
