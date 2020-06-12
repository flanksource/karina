package canary

import (
	"fmt"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "platform-system"

// Deploy deploys the canary-checker into the monitoring namespace
func Deploy(p *platform.Platform) error {
	if p.CanaryChecker == nil || !p.CanaryChecker.Disabled {
		if err := p.DeleteSpecs(Namespace, "canary-checker.yaml"); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}
		return nil
	}

	if p.CanaryChecker.ConfigFile == "" {
		return fmt.Errorf("must specify canaryChecker.configFile")
	}

	cfg := files.SafeRead(p.CanaryChecker.ConfigFile)
	if cfg == "" {
		p.Errorf("failed to read file: %v", p.CanaryChecker.ConfigFile)
	}

	if err := p.CreateOrUpdateConfigMap("canary-config", Namespace, map[string]string{
		"canary-config.yaml": cfg,
	}); err != nil {
		return err
	}

	return p.ApplySpecs(Namespace, "canary-checker.yaml")
}
