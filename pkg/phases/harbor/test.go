package harbor

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/net"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Harbor == nil || p.Harbor.Disabled {
		test.Skipf("Harbor", "Harbor is not configured")
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "harbor", test)
	health := fmt.Sprintf("%s/api/health", p.Harbor.URL)
	log.Infof("Checking %s\n", health)

	data, err := net.GET(health)
	if err != nil {
		test.Failf("Harbor", "Failed to get status from %s %s", health, err)
		return
	}
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		test.Failf("Harbor", "Failed to unmarshal json %v", err)
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

type Status struct {
	Status     string `json:"status"`
	Components []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"components"`
}
