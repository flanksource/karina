package kuberesourcereport

import (
	"fmt"
	"github.com/flanksource/commons/files"
	"os"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const Namespace = constants.PlatformSystem

func Install(p *platform.Platform) error {
	if p.KubeResourceReport == nil {
		p.KubeResourceReport = &types.KubeResourceReport{}
		if err := p.DeleteSpecs(Namespace, "kube-resource-report.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if p.KubeResourceReport.Disabled {

		if err := p.DeleteSpecs(Namespace, "kube-resource-report.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if p.KubeResourceReport.CustomCostFile != ""{
		_, err := os.Stat(p.KubeResourceReport.CustomCostFile)
		if err != nil {
			return fmt.Errorf("custom cost file %v not found: %v", p.KubeResourceReport.CustomCostFile,err)
		}
		data := files.SafeRead(p.KubeResourceReport.CustomCostFile)
		if data == "" {
			return fmt.Errorf("custom cost file %v is empty", p.KubeResourceReport.CustomCostFile)
		}
		p.CreateOrUpdateConfigMap("kube-resource-report",Namespace,
			map[string]string{
				"pricing.csv": data,
			})
	}

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
