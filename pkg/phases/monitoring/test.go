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
	"net/http"
	"time"
)

const (
	testMetricName  = "test_metric"
	testJobName     = "test"
	waitTimeSeconds = 0
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "monitoring", test)
}

func TestThanos(p *platform.Platform, test *console.TestResults, args []string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	pushGatewayHost := args[0]
	thanosHost := args[1]
	pusher := pushMetric(pushGatewayHost)
	err := pusher.Add()
    if err != nil {
		test.Failf("Thanos: client", "Failed to inject metric to client Prometheus via Pushgateway %v", err)
		return
	} else {
		test.Passf("Thanos client", "Metric successfully injected into the client Prometheus. Waiting to receive it in observability cluster.")
	}
	log.Tracef("Waiting for metric")
	time.Sleep(time.Second * waitTimeSeconds)
	metric, err := pullMetric(thanosHost)
    if err != nil {
		test.Failf("Thanos observability", "Failed to pull metric in Observability cluster %v", err)
	} else {
		log.Tracef("Got metric %v", metric)

		if metric.String() != "" {
			test.Passf("Thanos observability", "Got test metric successfully in Observability cluster")
			err = pusher.Delete()
			log.Info("Test metric deleted from pushgateway")
		} else {
			test.Failf("Thanos observability", "Failed to get test metric in Observability cluster")
		}
	}
	if err != nil {
		log.Warn("Failed to delete test metric from pushgateway")
		return
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
		RoundTripper: nil,
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
		fmt.Errorf("pullMetric: failed to ")
	}
	return value, err
}
