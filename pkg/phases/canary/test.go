package canary

import (
	"crypto/tls"
	"net/http"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testName = "canary"

func TestCanary(p *platform.Platform, test *console.TestResults) {
	if !p.Canary.Enabled {
		test.Skipf(testName, "canary is not enabled")
		return
	}
	client, err := p.GetClientset()
	if err != nil {
		test.Failf(testName, "couldn't get clientset: %v", err)
		return
	}

	ingress, err := client.ExtensionsV1beta1().Ingresses("monitoring").Get("canary", v1.GetOptions{})
	if err != nil {
		test.Failf(testName, "couldn't get ingress: %v", err)
		return
	}
	host := ingress.Spec.Rules[0].Host

	url := "http://" + host + "/metrics"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(url)
	if err != nil {
		test.Failf(testName, "couldn't GET %v : %v", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.Status != "200 OK" {
		test.Failf(testName, "metric status was not OK : %v", resp.Status)
		return
	}
	test.Passf(testName, "canary-checker metrics page up")
}
