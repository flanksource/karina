package templateoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "platform-system"
	Name      = "template-operator-controller-manager"
)

func Install(p *platform.Platform) error {
	_ = installNamespaceConfigurator(p)
	if p.TemplateOperator.IsDisabled() {
		return p.DeleteSpecs(Namespace, "template-operator.yaml")
	}

	return p.ApplySpecs(Namespace, "template-operator.yaml")
}

func installNamespaceConfigurator(platform *platform.Platform) error {
	if platform.NamespaceConfigurator == nil || platform.NamespaceConfigurator.Disabled {
		return platform.DeleteSpecs("", "namespace-configurator.yaml")
	}
	return platform.ApplySpecs("", "namespace-configurator.yaml")
}
