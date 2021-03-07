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
	TestKommonsTemplate(p, test)
}

func TestKommonsTemplate(platform *platform.Platform, test *console.TestResults) {
	testName := "kommons-template"

	templateResult, err := platform.TemplateText(`{{  kget "cm/quack/quack-config" "data.domain"  }}`)
	if err != nil {
		test.Failf(testName, "failed to template: %v", err)
	}
	if templateResult != platform.Domain {
		test.Failf(testName, "expected templated value to equal %s, got %s", platform.Domain, templateResult)
	} else {
		test.Passf(testName, "kget pulled cm/quack/quack-config successfully")
	}
}
