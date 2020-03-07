package stubs

import (
	"crypto/tls"
	"net/http"

	"github.com/flanksource/commons/console"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "minio", test)
	k8s.TestNamespace(client, "ldap", test)

	net := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	resp, err := net.Get("https://" + p.S3.GetExternalEndpoint())
	// 200 or 403 resoonse from minio is fine, 503 is not.
	if err != nil {
		test.Failf("minio", "minio GET / - %v", err)
	} else if resp.StatusCode == 200 || resp.StatusCode == 403 {
		test.Passf("minio", "minio GET /")
	} else {
		test.Failf("minio", "minio GET / - %v", resp.StatusCode)
	}

}
