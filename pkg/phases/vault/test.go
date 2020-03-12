package vault

import (
	"github.com/flanksource/commons/console"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Vault == nil || p.Vault.Disabled {
		test.Skipf("vault", "Vault is disabled")
		return
	}

	client, _ := p.GetClientset()
	k8s.TestNamespace(client, Namespace, test)
}
