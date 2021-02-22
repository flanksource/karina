package canary

import (
	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy deploys the canary-checker into the monitoring namespace
func Deploy(p *platform.Platform) error {
	if p.CanaryChecker == nil || p.CanaryChecker.Disabled {
		if err := p.DeleteSpecs(v1.NamespaceAll, "canary-checker.yaml"); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}
		if err := p.DeleteSpecs(v1.NamespaceAll, "canary-checker-alerts.yaml.raw"); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}
		return nil
	}

	if p.CanaryChecker.Version == "" {
		p.CanaryChecker.Version = "v0.11.2"
	}
	if err := p.ApplySpecs(v1.NamespaceAll, "canary-checker.yaml"); err != nil {
		return err
	}
	return p.ApplySpecs(v1.NamespaceAll, "canary-checker-alerts.yaml.raw")
}
