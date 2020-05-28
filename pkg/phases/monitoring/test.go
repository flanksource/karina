package monitoring

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

const (
	testMetricName = "test_metric"
	testJobName    = "test"
	critical       = "critical"
	warning        = "warning"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "monitoring", test)
}

func TestThanos(p *platform.Platform, test *console.TestResults) {
	if p.Thanos == nil || p.Thanos.Disabled {
		test.Skipf("thanos", "thanos is disabled")
		return
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	pushGatewayHost := "pushgateway." + p.Domain
	thanosHost := fmt.Sprintf("https://thanos.%s", p.Domain)

	if p.Thanos.Mode != "observability" {
		if p.Thanos.E2E.Server == "" {
			test.Skipf("thanos", "Must specify a thanos server under e2e.server in client mode")
			return
		}
		thanosHost = p.Thanos.E2E.Server
	}

	pusher := pushMetric(pushGatewayHost)
	err := pusher.Add()
	if err != nil {
		test.Failf("Thanos: client", "Failed to inject metric to client Prometheus via Pushgateway %v", err)
		return
	}
	test.Passf("Thanos client", "Metric successfully injected into the client Prometheus. Waiting to receive it in observability cluster.")
	retries := 12
	for {
		if retries == 0 {
			test.Failf("Thanos observability", "Failed to get test metric in Observability cluster")
			break
		}
		metric, err := pullMetric(thanosHost)
		if err != nil {
			test.Failf("Thanos observability", "Failed to pull metric in Observability cluster %v", err)
		} else {
			test.Tracef("Got metric %v", metric)
			if metric.String() != "" {
				test.Passf("Thanos observability", "Got test metric successfully in Observability cluster")
				_ = pusher.Delete()
				test.Infof("Test metric deleted from pushgateway")
				break
			} else {
				time.Sleep(time.Second * 5)
				test.Tracef("Retrying")
				retries--
			}
		}
		if err != nil {
			test.Warnf("Failed to delete test metric from pushgateway")
			break
		}
	}
}

func TestPrometheus(p *platform.Platform, test *console.TestResults) {
	if p.Monitoring == nil || p.Monitoring.Disabled {
		test.Skipf("prometheus", "monitoring is not configured or enabled")
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
	promAPI := v1.NewAPI(client)
	targets, err := promAPI.Targets(context.Background())
	if err != nil {
		test.Failf("prometheus", "Failed to get targets: %v", err)
		return
	}
	if targets.Active == nil {
		test.Failf("NoActiveTargets", "No active targets found in Prometheus")
		return
	}
	for _, activeTarget := range targets.Active {
		targetEndpointName := activeTarget.DiscoveredLabels["__meta_kubernetes_endpoints_name"]
		targetEndpointAddress := activeTarget.DiscoveredLabels["__address__"]
		if activeTarget.Health == "down" {
			test.Failf(targetEndpointName, "%s (%s) endpoint is down\n %s", targetEndpointName, targetEndpointAddress, activeTarget.LastError)
		} else {
			test.Passf(targetEndpointName, "%s (%s) endpoint is up", targetEndpointName, targetEndpointAddress)
		}
	}
	alerts, err := promAPI.Alerts(context.Background())
	if err != nil {
		test.Errorf("pullMetric: failed to get alerts")
		return
	}
	if alerts.Alerts == nil {
		test.Failf("prometheus", "Watchdog alert should be firing")
		return
	}
	alertLevel := p.Monitoring.E2E.MinAlertLevel
	if alertLevel == "" {
		alertLevel = critical
	}
	for _, alert := range alerts.Alerts {
		alertname := string(alert.Labels["alertname"])
		severity := string(alert.Labels["severity"])
		if alertname == "Watchdog" {
			continue
		}
		if alertLevel == critical && severity != critical {
			continue
		}
		if alertLevel == warning && severity != critical && severity != warning {
			continue
		}
		if alert.State == "firing" {
			test.Failf(alertname, "%s alert is firing %s", alertname, alert.Labels)
		}
	}
}

func pushMetric(pushGatewayHost string) *push.Pusher {
	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: testMetricName,
		Help: "This metric is created from Thanos e2e tests. You can ignore it.",
	})
	registry := prometheus.NewRegistry()
	registry.MustRegister(metric)
	metric.Add(1)
	pusher := push.New(pushGatewayHost, testJobName).Gatherer(registry)
	return pusher
}

func pullMetric(thanosHost string) (model.Value, error) {
	client, err := api.NewClient(api.Config{
		Address:      thanosHost,
		RoundTripper: http.DefaultTransport,
	})
	if err != nil {
		return nil, fmt.Errorf("pullMetric: failed to get api client to connect to Thanos: %s", err)
	}
	promAPI := v1.NewAPI(client)
	value, warn, err := promAPI.Query(context.Background(), testMetricName, time.Now())
	if len(warn) != 0 {
		log.Tracef("Got warnings: %s", warn)
	}
	if err != nil {
		log.Errorf("pullMetric: failed to pull metrics")
	}
	return value, err
}
