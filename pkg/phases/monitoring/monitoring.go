package monitoring

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(p *platform.Platform) error {
	if p.Monitoring == nil || p.Monitoring.Disabled {
		return nil
	}
	kubectl := p.GetKubectl()
	dir, err := p.TemplateDir("monitoring/")
	if err != nil {
		return err
	}
	return kubectl("kustomize %s | .bin/kubectl apply -f -", dir)
}
