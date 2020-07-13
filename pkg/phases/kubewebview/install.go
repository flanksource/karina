package kubewebview

import (
	"fmt"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const Namespace = constants.PlatformSystem

func Install(p *platform.Platform) error {
	if p.KubeWebView == nil || p.KubeWebView.Disabled {
		p.KubeWebView = &types.KubeWebView{}
		if err := p.DeleteSpecs(Namespace, "kube-web-view.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	return p.ApplySpecs(Namespace, "kube-web-view.yaml")
}
