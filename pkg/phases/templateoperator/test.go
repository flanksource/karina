package templateoperator

import (
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.TemplateOperator.IsDisabled() {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf(Name, "Could not connect to Platform client: %v", err)
		return
	}
	if err := p.WaitForDeployment(Namespace, Name, 60*time.Second); err != nil {
		test.Failf(Name, "template-operator did not become ready")
		return
	}
	kommons.TestDeploy(client, Namespace, Name, test)
	TestKommonsTemplate(p, test)
}

func TestKommonsTemplate(platform *platform.Platform, test *console.TestResults) {
	testName := "kommons-template"

	templateResult, err := platform.TemplateText(`{{  kget "cm/quack/quack-config" "data.domain"  }}`)
	if err != nil {
		test.Failf(testName, "failed to template: %v", err)
		return
	}
	if templateResult != platform.Domain {
		test.Failf(testName, "expected templated value to equal %s, got %s", platform.Domain, templateResult)
	} else {
		test.Passf(testName, "kget pulled cm/quack/quack-config successfully")
	}
}
