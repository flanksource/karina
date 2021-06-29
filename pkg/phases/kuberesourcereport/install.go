package kuberesourcereport

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace = constants.PlatformSystem
	Group     = "system:reporting"
	User      = "kube-resource-report"
	// if a region is specified (it's labels is not set) then
	// kube-resource-report uses this region.
	// from https://github.com/hjacobs/kube-resource-report/blob/cd43749cd191e17f62a63f9f74757fcad487c181/kube_resource_report/query.py#L232
	// We use this to avoid forcing the user to specify a
	// usually unused region in the
	// region,instancetype,cost in the
	// platform YAML file.
	DefaultRegion  = "unknown"
	ClusterConfigs = "kube-resource-report-clusters"
)

func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil {
		p.KubeResourceReport = &types.KubeResourceReport{} // this sets p.KubeResourceReport.Disabled to false
		p.KubeResourceReport.Disabled = true
	}

	if p.DryRun && !p.KubeResourceReport.Disabled {
		return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
	} else if p.DryRun {
		return nil
	}

	if p.KubeResourceReport.Disabled {
		// remove the secret containing access information to external clusters
		cs, err := p.GetClientset()
		if err != nil {
			return fmt.Errorf("failed to get clientset for cluster: %v", err)
		}

		if p.HasSecret(Namespace, ClusterConfigs) {
			if err := cs.CoreV1().Secrets(Namespace).Delete(context.TODO(), ClusterConfigs, metav1.DeleteOptions{}); err != nil {
				return err
			}
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
	kubeConfig, err := kommons.CreateMultiKubeConfig(ca, p.KubeResourceReport.ExternalClusters, Group, User, 2*356*24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to generate kubeconfig for multi-cluster access: %v", err)
	}
	err = p.CreateOrUpdateSecret("kube-resource-report-clusters", Namespace, map[string][]byte{
		"config": kubeConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to generate kubeconfig secret for multi-cluster access: %v", err)
	}
	p.Infof("Created external cluster access secret")

	customCostGeneratedData := ""
	for label, value := range p.KubeResourceReport.Costs {
		p.Infof("Reading custom cost label: %v, %.3f", label, value)
		if !strings.Contains(label, ",") {
			//kube-resource-report does not like spaces
			newRow := fmt.Sprintf("%v,%v,%d\n", DefaultRegion, label, value) // nolint: govet, staticcheck
			customCostGeneratedData = customCostGeneratedData + newRow
			p.Debugf("Adding custom cost label: %v", newRow)
		} else {
			split := strings.SplitAfterN(label, ",", 2)
			region := split[0]
			label := split[1]
			//split string contains the , so not added again
			//kube-resource-report does not like spaces
			newRow := fmt.Sprintf("%v%v,%d\n", region, label, value) // nolint: govet, staticcheck
			customCostGeneratedData = customCostGeneratedData + newRow
			p.Debugf("Adding custom cost label: %v", newRow)
		}
	}

	customCostReadData := ""
	if p.KubeResourceReport.CostsFile != "" {
		_, err := os.Stat(p.KubeResourceReport.CostsFile)
		if err != nil {
			return fmt.Errorf("custom cost file %v not found: %v", p.KubeResourceReport.CostsFile, err)
		}
		customCostReadData = files.SafeRead(p.KubeResourceReport.CostsFile)
		if customCostReadData == "" {
			return fmt.Errorf("custom cost file %v is empty", p.KubeResourceReport.CostsFile)
		}
	}
	if len(customCostReadData) > 0 || len(customCostGeneratedData) > 0 {
		err = p.CreateOrUpdateConfigMap("kube-resource-report", Namespace,
			map[string]string{
				"pricing.csv": customCostGeneratedData + customCostReadData,
			})
		if err != nil {
			return fmt.Errorf("custom cost configmap creation failed: %v", err)
		}
	}

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
