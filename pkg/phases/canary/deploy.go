package canary

import (
	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var specs = []string{"canary-checker.yaml", "canary-checker-monitoring.yaml.raw"}

// Deploy deploys the canary-checker into the monitoring namespace
func Deploy(p *platform.Platform) error {
	if p.CanaryChecker.IsDisabled() {
		if err := p.DeleteSpecs(v1.NamespaceAll, specs...); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}

		return nil
	}

	err := p.DeleteByKind("Ingress", "platform-system", "canary-checker")
	if err != nil {
		p.Warnf("failed to delete old ingress: ", err)
	}

	return p.ApplySpecs(v1.NamespaceAll, specs...)
}
