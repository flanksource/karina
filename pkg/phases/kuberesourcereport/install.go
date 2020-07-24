package kuberesourcereport

import (
	"fmt"
	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/k8s"
	"net/url"
	"time"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	//"k8s.io/client-go/pkg/a"
	//_ "k8s.io/client-go/pkg/api/install"
	//_ "k8s.io/client-go/pkg/apis/extensions/install"
)

const (
	Namespace = constants.PlatformSystem
	Group     = "system:reporting"
	User      = "kube-resource-report"
)

func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil || p.KubeResourceReport.Disabled {
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
		clusters := map[string]string{}
		//	p.Name: "https://kubernetes.default.svc",
		//}
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
			template, err := p.Template("kube-resource-report-rbac.yaml", "manifests")
			if err != nil {
				return fmt.Errorf("applySpecs: failed to template manifests: %v", err)
			}
			p.Logger.Infof("template is \n%v", template)

			client.ApplyText(Namespace, template)
			if err != nil {
				fmt.Errorf("error applying external cluster security manifest %v", err)
			}
			clusters[name] = apiEndpoint

		}

		kubeConfig, err := k8s.CreateMultiKubeConfig(ca, clusters, Group, User, 24*7*time.Hour)
		p.Logger.Infof("kubeconfig file is:\n%v", string(kubeConfig))

		p.CreateOrUpdateSecret("kube-resource-report-clusters", Namespace, map[string][]byte{
			"config": []byte(kubeConfig),
		})

	}

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
