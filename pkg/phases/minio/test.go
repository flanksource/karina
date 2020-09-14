package minio

import (
	"crypto/tls"
	"net/http"

	"github.com/flanksource/commons/console"

	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/k8s/proxy"
	"github.com/flanksource/karina/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Minio.IsDisabled() {
		return
	}
	client, _ := p.GetClientset()

	k8s.TestNamespace(client, Namespace, test)

	dialer, _ := p.GetProxyDialer(proxy.Proxy{
		Namespace:    Namespace,
		Kind:         "pods",
		ResourceName: "minio-0",
		Port:         9000,
	})
	net := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	netClient := http.Client{Transport: net}
	resp, err := netClient.Get("http://host")
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() // nolint: errcheck
	}
	// 200 or 403 response from minio is fine, 503 is not.
	if err != nil {
		test.Failf("minio", "minio GET / - %v", err)
	} else if resp.StatusCode == 200 || resp.StatusCode == 403 {
		test.Passf("minio", "minio GET /")
	} else {
		test.Failf("minio", "minio GET / - %v", resp.StatusCode)
	}
}
