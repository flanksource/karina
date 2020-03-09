package postgresOperator

import (
	"github.com/flanksource/commons/console"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.PostgresOperator == nil || p.PostgresOperator.Disabled {
		test.Skipf("postgres-operator", "Postgres operator is disabled")
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, Namespace, test)
}
