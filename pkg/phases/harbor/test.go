package harbor

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Harbor.IsDisabled() {
		return
	}
	client, _ := p.GetClientset()
	kommons.TestNamespace(client, "harbor", test)
	harbor, err := NewClient(p)
	if err != nil {
		test.Failf(Namespace, "failed to get harbor client: %v", err)
		return
	}
	var status *Status
	status, err = harbor.GetStatus()
	if err != nil {
		test.Failf("Harbor", "Failed to get harbor status %v", err)
		return
	}

	for _, component := range status.Components {
		if component.Status == "healthy" {
			test.Passf("Harbor", component.Name)
		} else {
			test.Failf("Harbor", "%s is %s", component.Name, component.Status)
		}
	}
}
