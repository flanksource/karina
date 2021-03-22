package dashboard

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = constants.KubeSystem
)

func Install(p *platform.Platform) error {
	if p.Dashboard.Disabled {
		return p.DeleteSpecs(Namespace, "k8s-dashboard.yaml")
	}

	return p.ApplySpecs(Namespace, "k8s-dashboard.yaml")
}
