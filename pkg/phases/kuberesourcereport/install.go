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
)

const Namespace = constants.PlatformSystem

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
		ep, err := p.GetAPIEndpoint()
		if err != nil {
			fmt.Errorf("Unable to get cluster endpoint %v", err)
		}
		clusters := map[string]string{
			p.Name: "https://"+ep+":6443",
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


			client := k8s.GetExternalClient(name,u.Hostname(), ca)

			mast, err := client.GetMasterNode()
			if err != nil {
				fmt.Errorf("failed getting clientset %v", err)
			}
			p.Logger.Infof("master is %v",mast)
			clusters[name] = apiEndpoint

			//kubeConfig, err := k8s.CreateKubeConfig(name, ca, u.Hostname(), "system:masters", "admin", 24*7*time.Hour)
			//if err != nil {
			//	p.Logger.Infof("error getting kubeconfig:\n%v", err)
			//}
			//p.Logger.Infof("kubeconfig file is:\n%v",string(kubeConfig))
		}

		kubeConfig, err := k8s.CreateMultiKubeConfig(ca, clusters, "system:masters", "kubernetes-admin", 24*7*time.Hour)
		p.Logger.Infof("kubeconfig file is:\n%v",string(kubeConfig))

		p.CreateOrUpdateSecret("kube-resource-report-clusters",Namespace, map[string][]byte {
			"config" : []byte(kubeConfig),
			})

	}


	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
