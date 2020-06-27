package canary

import (
	"fmt"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy deploys the canary-checker into the monitoring namespace
func Deploy(p *platform.Platform) error {
	if p.CanaryChecker == nil || p.CanaryChecker.Disabled {
		if err := p.DeleteSpecs(v1.NamespaceAll, "canary-checker.yaml"); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}
		return nil
	}

	if p.CanaryChecker.Version == "" {
		p.CanaryChecker.Version = "v0.10.1"
	}

	if p.CanaryChecker.Interval == 0 {
		p.Warnf("Canary Checker interval is missing, default is 1 minute")
		p.CanaryChecker.Interval = 60
	}

	if p.CanaryChecker.ConfigFile == "" {
		return fmt.Errorf("must specify canaryChecker.configFile")
	}

	cfg := files.SafeRead(p.CanaryChecker.ConfigFile)
	if cfg == "" {
		p.Errorf("failed to read file: %v", p.CanaryChecker.ConfigFile)
	}

	if err := p.CreateOrUpdateConfigMap("canary-config", constants.PlatformSystem, map[string]string{
		"canary-config.yaml": cfg,
	}); err != nil {
		return err
	}

	return p.ApplySpecs(v1.NamespaceAll, "canary-checker.yaml")
}
