package monitoring

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

const (
	testMetricName  = "test_metric"
	testJobName     = "test"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "monitoring", test)
}

func TestThanos(p *platform.Platform, test *console.TestResults, _ []string, cmd *cobra.Command) {
	if p.Thanos == nil {
		log.Fatalf("thanos is disabled")
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	flags := cmd.Flags()
	pushGatewayHost, _ := flags.GetString("pushgateway")
	thanosHost, _ := flags.GetString("thanos")
	if thanosHost == "" {
		if p.Thanos.Mode != "observability" {
			log.Fatalf("please specify --thanos flag in client mode")
		} else {
			thanosHost = fmt.Sprintf("https://thanos.%s", p.Domain)
		}
	}
	if pushGatewayHost == "" {
		pushGatewayHost = fmt.Sprintf("pushgateway.%s", p.Domain)
	}
	pusher := pushMetric(pushGatewayHost)
	err := pusher.Add()
    if err != nil {
		test.Failf("Thanos: client", "Failed to inject metric to client Prometheus via Pushgateway %v", err)
		return
	} else {
		test.Passf("Thanos client", "Metric successfully injected into the client Prometheus. Waiting to receive it in observability cluster.")
	}
	log.Tracef("Waiting for metric")
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
			log.Tracef("Got metric %v", metric)
			if metric.String() != "" {
				test.Passf("Thanos observability", "Got test metric successfully in Observability cluster")
				err = pusher.Delete()
				log.Info("Test metric deleted from pushgateway")
				break
			} else {
				time.Sleep(time.Second * 5)
				log.Trace("Retrying")
				retries -= 1
			}
		}
		if err != nil {
			log.Warn("Failed to delete test metric from pushgateway")
			break
		}
	}

}

func TestPrometheus(p *platform.Platform, test *console.TestResults, _ []string, cmd *cobra.Command) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client, err := api.NewClient(api.Config{
		Address:      fmt.Sprintf("https://prometheus.%s", p.Domain),
		RoundTripper: http.DefaultTransport,
	})
	if err != nil {
		log.Fatal("Failed to get client to connect to Prometheus")
	}
	promApi := v1.NewAPI(client)
	targets, err := promApi.Targets(context.Background())
	if err != nil {
		log.Fatal("Failed to get targets")
	}
	if targets.Active == nil {
		test.Failf("NoActiveTargets", "No active targets found in Prometheus")
		return
	}
	for _, activeTarget := range targets.Active {
		targetEndpointName := activeTarget.DiscoveredLabels["__meta_kubernetes_endpoints_name"]
		targetEndpointAddress := activeTarget.DiscoveredLabels["__address__"]
		log.Tracef("Testing endpoint: %s", targetEndpointName)
		log.Trace(activeTarget)
		if activeTarget.Health == "down" {
			test.Failf(targetEndpointName, "%s (%s) endpoint is down\n %s", targetEndpointName, targetEndpointAddress, activeTarget.LastError)
		} else {
			test.Passf(targetEndpointName, "%s (%s) endpoint is up", targetEndpointName, targetEndpointAddress)
		}
	}
	alerts, err := promApi.Alerts(context.Background())
	if err != nil {
		log.Errorf("pullMetric: failed to get alerts")
		return
	}
	if alerts.Alerts == nil {
		log.Infof("No alerts")
		return
	}
	for _, alert := range alerts.Alerts {
		alertname := string(alert.Labels["alertname"])
		if alert.State == "firing" {
			test.Failf(alertname, "%s alert is firing\n %s", alertname, alert.Labels)
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
	promApi := v1.NewAPI(client)
	value, warn, err := promApi.Query(context.Background(), testMetricName, time.Now())
	if len(warn) != 0 {
		log.Tracef("Got warnings: %s", warn)
	}
	if err != nil {
		fmt.Errorf("pullMetric: failed to pull metrics")
	}
	return value, err
}
