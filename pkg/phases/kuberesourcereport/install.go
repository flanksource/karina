package kuberesourcereport

import (
	"fmt"
	"net/url"
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

// Install deploys and configures kube-resource-report to the platform.
// If external clusters are specified they are configured
// to allow access by creating a ClusterRole and ClusterRoleBinding.
// If the configuration indicates that kube-resource-report is disabled
// then it will be uninstalled and external configurations undone.
func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil {
		p.KubeResourceReport = &types.KubeResourceReport{}
		return p.DeleteSpecs(Namespace, "kube-resource-report.yaml")
	}
	if p.KubeResourceReport.Disabled {
		// if external clusters were configured we attempt to remove the configuration first
		if len(p.KubeResourceReport.ExternalClusters) > 0 {
			err := removeExternalClusterRBAC(p)
			if err != nil {
				p.Warnf("failed to remove external cluster RBAC configs: %v", err)
				// keep going - failure to remove access doesn't stop the uninstall
			}
			// remove the secret containing access information to external clusters
			cs, err := p.GetClientset()
			if err != nil || cs == nil {
				p.Warnf("failed to get clientset for cluster: %v", err)
			} else {
				err = cs.CoreV1().Secrets(Namespace).Delete("kube-resource-report-clusters", &metav1.DeleteOptions{})
				if err != nil {
					p.Warnf("failed to get clientset for cluster: %v", err)
				}
			}
		}
		return p.DeleteSpecs(Namespace, "kube-resource-report.yaml")
	}

	// make sure the namespace exists
	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	// if external clusters are specified we configure them to allow access
	if len(p.KubeResourceReport.ExternalClusters) > 0 {
		// we use our own root CA for ALL cluster accesses
		ca, err := ca.ReadCA(p.CA)
		if err != nil {
			p.Errorf("Unable to get root CA %v", err)
			// keep going - failure to configure access doesn't stop the install
		}
		clusters, err := addExternalClusterRBAC(p)
		if err != nil {
			p.Warnf("failed to add external cluster RBAC configs: %v", err)
			// keep going - failure to configure access doesn't stop the install
		}
		// create a secret containing a kubeconfig file that allows access to
		// this cluster via user/cert as well as the given external clusters
		kubeConfig, err := k8s.CreateMultiKubeConfig(ca, *clusters, Group, User, 24*7*time.Hour)
		if err != nil {
			p.Warnf("failed to generate kubeconfig for multi-cluster access: %v", err)
			// keep going - failure to configure access doesn't stop the install
		}
		if p.PlatformConfig.Trace {
			p.Infof("kubeconfig file is:\n%v", string(kubeConfig))
		}
		err = p.CreateOrUpdateSecret("kube-resource-report-clusters", Namespace, map[string][]byte{
			"config": kubeConfig,
		})
		if err != nil {
			p.Warnf("failed to generate kubeconfig secret for multi-cluster access: %v", err)
			// keep going - failure to configure access doesn't stop the install
		}
	}
	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}

// addExternalClusterRBAC configures a ClusterRole and ClusterRoleBinding for
// each specified external cluster using the template kube-resource-report-external-rbac.yaml
// it returns a map of configured cluster, cluster API endpoints
func addExternalClusterRBAC(p *platform.Platform) (*map[string]string, error) {
	if len(p.KubeResourceReport.ExternalClusters) < 1 {
		return nil, fmt.Errorf("no external clusters configured")
	}
	// we use our own root CA for ALL cluster accesses
	ca, err := ca.ReadCA(p.CA)
	if err != nil {
		return nil, fmt.Errorf("unable to get root CA %v", err)
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
			p.Errorf("Unable to parse external cluster endpoint URL: %v", apiEndpoint)
			continue
			// failing to add this external cluster - try the next one
		}
		if u.Port() != "6443" {
			p.Errorf("Only port 6443 supported for external cluster endpoint URLs: %v", apiEndpoint)
			continue
			// failing to add this external cluster - try the next one
		}
		p.Debugf("External endpoint host: %v", u.Hostname())

		client := k8s.GetExternalClient(p.Logger, name, u.Hostname(), ca)

		// create rbac settings for remote user
		template, err := p.Template("kube-resource-report-external-rbac.yaml", "manifests")
		if err != nil {
			p.Errorf("applySpecs: failed to template manifests: %v", err)
			continue
			// failing to add this external cluster - try the next one
		}
		if p.PlatformConfig.Trace {
			p.Infof("template is: \n%v", template)
		}

		err = client.ApplyText(Namespace, template)
		if err != nil {
			p.Errorf("error applying external cluster security manifest %v", err)
			continue
			// failing to add this external cluster - try the next one
		}
		// if the cluster was configured we return it in the result map
		clusters[name] = apiEndpoint
	}
	return &clusters, nil
}

// removeExternalClusterRBAC removes a ClusterRole and ClusterRoleBinding for
// each specified external cluster using the template kube-resource-report-external-rbac.yaml
func removeExternalClusterRBAC(p *platform.Platform) error {
	if len(p.KubeResourceReport.ExternalClusters) < 1 {
		return fmt.Errorf("no external clusters configured")
	}
	// we use our own root CA for ALL cluster accesses
	ca, err := ca.ReadCA(p.CA)
	if err != nil {
		return fmt.Errorf("unable to get root CA %v", err)
	}

	for name, apiEndpoint := range p.KubeResourceReport.ExternalClusters {
		p.Infof("Removing RBAC from external cluster %v with endpoint: %v from kube-resource-report.", name, apiEndpoint)

		u, err := url.Parse(apiEndpoint)
		if err != nil {
			p.Errorf("Unable to parse external cluster endpoint URL: %v", apiEndpoint)
			continue
			// failing to add this external cluster - try the next one
		}
		if u.Port() != "6443" {
			p.Errorf("Only port 6443 supported for external cluster endpoint URLs: %v", apiEndpoint)
			continue
			// failing to add this external cluster - try the next one
		}

		p.Debugf("External endpoint host: %v", u.Hostname())

		client := k8s.GetExternalClient(p.Logger, name, u.Hostname(), ca)

		// create rbac settings for remote user
		template, err := p.Template("kube-resource-report-external-rbac.yaml", "manifests")
		if err != nil {
			p.Errorf("applySpecs: failed to template manifests: %v", err)
			continue
			// failing to add this external cluster - try the next one
		}
		if p.PlatformConfig.Trace {
			p.Infof("template is: \n%v", template)
		}

		err = client.DeleteText(Namespace, template)
		if err != nil {
			p.Errorf("error deleting external cluster security manifest %v", err)
		}
	}
	return nil
}
