package kuberesourcereport

import (
	"fmt"
	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"time"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	//"k8s.io/client-go/pkg/a"
	//_ "k8s.io/client-go/pkg/api/install"
	//_ "k8s.io/client-go/pkg/apis/extensions/install"
)

const (
	Namespace = constants.PlatformSystem
	Group     = "system:reporting"
	User      = "kube-resource-report"
)

// Install deploys kube-resource-report to the platform
func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil || p.KubeResourceReport.Disabled {

		if len(p.KubeResourceReport.ExternalClusters) > 0 {
			err := RemoveExternalClusterRBAC(p)
			if err != nil {
				p.Warnf("failed to remove external cluster RBAC configs: %v", err)
			}

			cs, _ := p.GetClientset()
			cs.CoreV1().Secrets(Namespace).Delete("kube-resource-report-clusters", &metav1.DeleteOptions{})

		}
		p.KubeResourceReport = &types.KubeResourceReport{}
		if err := p.DeleteSpecs(Namespace, "kube-resource-report.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}


		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if len(p.KubeResourceReport.ExternalClusters) > 0 {
		// we use our own root CA for ALL cluster accesses
		ca, err := ca.ReadCA(p.CA)
		if err != nil {
			fmt.Errorf("Unable to get root CA %v", err)
		}
		clusters, err := AddExternalClusterRBAC(p)
		if err != nil {
			p.Warnf("failed to add external cluster RBAC configs: %v", err)
		}
		kubeConfig, err := k8s.CreateMultiKubeConfig(ca, *clusters, Group, User, 24*7*time.Hour)
		p.Logger.Infof("kubeconfig file is:\n%v", string(kubeConfig))

		p.CreateOrUpdateSecret("kube-resource-report-clusters", Namespace, map[string][]byte{
			"config": []byte(kubeConfig),
		})

	}

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}

func AddExternalClusterRBAC(p *platform.Platform) (*map[string]string, error) {
	if len(p.KubeResourceReport.ExternalClusters)<1 {
		return nil, fmt.Errorf("no external clusters configured.")
	}
	// we use our own root CA for ALL cluster accesses
	ca, err := ca.ReadCA(p.CA)
	if err != nil {
		fmt.Errorf("Unable to get root CA %v", err)
	}

	clusters := map[string]string{
		// kube-resource-view can't use the service account to access it's own cluster
		// so we add user/cert access via the default internal API endpoint
		p.Name: "https://kubernetes.default",
	}
	for name, apiEndpoint := range p.KubeResourceReport.ExternalClusters {
		p.Logger.Infof("Adding external cluster %v with endpoint: %v to kube-resource-report.", name, apiEndpoint)

		u, err := url.Parse(apiEndpoint)
		if err != nil {
			fmt.Errorf("Unable to parse external cluster endpoint URL: %v", apiEndpoint)
		}
		if u.Port() != "6443" {
			fmt.Errorf("Only port 6443 supported for external cluster endpoint URLs: %v", apiEndpoint)
		}

		p.Logger.Debugf("External endpoint host: %v", u.Hostname())

		client := k8s.GetExternalClient(p.Logger, name, u.Hostname(), ca)

		mast, err := client.GetMasterNode()
		if err != nil {
			fmt.Errorf("failed getting clientset %v", err)
		}
		p.Logger.Infof("master is %v", mast)

		// create rbac settings for remote user
		template, err := p.Template("kube-resource-report-external-rbac.yaml", "manifests")
		if err != nil {
			return nil, fmt.Errorf("applySpecs: failed to template manifests: %v", err)
		}
		p.Logger.Infof("template is \n%v", template)

		client.ApplyText(Namespace, template)
		if err != nil {
			fmt.Errorf("error applying external cluster security manifest %v", err)
		}
		clusters[name] = apiEndpoint

	}
	return &clusters, nil
}

func RemoveExternalClusterRBAC(p *platform.Platform) error {
	if len(p.KubeResourceReport.ExternalClusters)<1 {
		return fmt.Errorf("no external clusters configured.")
	}
	// we use our own root CA for ALL cluster accesses
	ca, err := ca.ReadCA(p.CA)
	if err != nil {
		fmt.Errorf("Unable to get root CA %v", err)
	}

	for name, apiEndpoint := range p.KubeResourceReport.ExternalClusters {
		p.Logger.Infof("Removing RBAC from external cluster %v with endpoint: %v from kube-resource-report.", name, apiEndpoint)

		u, err := url.Parse(apiEndpoint)
		if err != nil {
			fmt.Errorf("Unable to parse external cluster endpoint URL: %v", apiEndpoint)
		}
		if u.Port() != "6443" {
			fmt.Errorf("Only port 6443 supported for external cluster endpoint URLs: %v", apiEndpoint)
		}

		p.Logger.Debugf("External endpoint host: %v", u.Hostname())

		client := k8s.GetExternalClient(p.Logger, name, u.Hostname(), ca)

		mast, err := client.GetMasterNode()
		if err != nil {
			fmt.Errorf("failed getting clientset %v", err)
		}
		p.Logger.Infof("master is %v", mast)

		// create rbac settings for remote user
		template, err := p.Template("kube-resource-report-external-rbac.yaml", "manifests")
		if err != nil {
			return fmt.Errorf("applySpecs: failed to template manifests: %v", err)
		}
		p.Logger.Infof("template is \n%v", template)

		err = client.DeleteText(Namespace, template)
		if err != nil {
			fmt.Errorf("error deleting external cluster security manifest %v", err)
		}

	}
	return nil
}
