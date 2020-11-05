package stubs

import (
	"crypto/tls"
	"net/http"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.Ldap.E2E.Mock {
		kommons.TestNamespace(client, "ldap", test)
	}
	if !p.S3.E2E.Minio {
		return
	}
	kommons.TestNamespace(client, "minio", test)

	net := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	resp, err := net.Get("https://" + p.S3.Endpoint)
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
