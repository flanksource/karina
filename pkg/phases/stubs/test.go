package stubs

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)


func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "minio", test)
	k8s.TestNamespace(client, "ldap", test)
}
