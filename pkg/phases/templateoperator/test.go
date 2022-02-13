package templateoperator

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.TemplateOperator.IsDisabled() {
		return
	}

	if err := p.WaitForDeployment(Namespace, Name, 60*time.Second); err != nil {
		test.Failf(Name, "template-operator did not become ready")
		return
	}
	test.Passf(Name, "template-operator is ready")
}
