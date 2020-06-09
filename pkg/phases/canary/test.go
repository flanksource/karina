package canary

import (
	"crypto/tls"
	"fmt"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
)

func TestCanary(p *platform.Platform, test *console.TestResults) {
	if !p.Canary.Enabled {
		test.Skipf("canary", "canary is not enabled")
		return
	}
	if p.Monitoring == nil || p.Monitoring.Disabled {
		test.Skipf("canary", "prometheus monitoring is not configured or enabled")
		return
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client, err := api.NewClient(api.Config{
		Address:      fmt.Sprintf("https://prometheus.%s", p.Domain),
		RoundTripper: http.DefaultTransport,
	})
	if err != nil {
		test.Failf("prometheus", "Failed to get client to connect to Prometheus")
		return
	}
	_ = v1.NewAPI(client)

}
