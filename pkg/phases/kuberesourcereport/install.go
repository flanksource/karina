package kuberesourcereport

import (
	"fmt"

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

	return p.ApplySpecs(Namespace, "kube-resource-report.yaml")
}
