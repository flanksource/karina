package harbor

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/commons/console"
	"github.com/moshloop/commons/net"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	defaults(p)
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "harbor", test)
	health := fmt.Sprintf("%s/api/health", p.Harbor.URL)
	log.Infof("Checking %s\n", health)

	data, err := net.GET(health)
	if err != nil {
		test.Failf("Failed to get status from %s %s", health, err)
		return
	}
	var status HarborStatus
	json.Unmarshal(data, &status)

	for _, component := range status.Components {
		if component.Status == "healthy" {
			test.Passf("Harbor", component.Name)
		} else {
			test.Failf("Harbor", "%s is %s", component.Name, component.Status)
		}
	}
}

type HarborStatus struct {
	Status     string `json:"status"`
	Components []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"components"`
}
