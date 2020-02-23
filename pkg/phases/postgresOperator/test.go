package postgresOperator

import (
	"github.com/flanksource/commons/console"
	"k8s.io/client-go/kubernetes"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func TestNamespace(p *platform.Platform, client kubernetes.Interface, test *console.TestResults) {
	k8s.TestNamespace(client, "postgres-operator", test)
}
