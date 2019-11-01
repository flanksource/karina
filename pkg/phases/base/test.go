package base

import (
	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "kube-system", test)
	k8s.TestNamespace(client, "ingress-nginx", test)
	k8s.TestNamespace(client, "local-path-storage", test)
	k8s.TestNamespace(client, "cert-manager", test)
}
