package canary

import (
	"fmt"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/platform"
)

// Deploy deploys the canary-checker into the Platform System namespace
func Deploy(p *platform.Platform) error {
	if !p.Canary.Enabled {
		if err := p.DeleteSpecs("monitoring", "canary.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	cfg := files.SafeRead(p.Canary.ConfigFile)
	if cfg == "" {
		p.Errorf("failed to read file: %v", p.Canary.ConfigFile)
	}

	if err := p.CreateOrUpdateConfigMap("canary-config", "monitoring", map[string]string{
		"ConfigName": cfg,
	}); err != nil {
		return fmt.Errorf("install: failed to create/update configmap: %v", err)
	}

	return p.ApplySpecs("monitoring", "canary.yaml")
}
