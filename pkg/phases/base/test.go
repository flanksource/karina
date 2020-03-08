package base

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	client, _ := platform.GetClientset()
	k8s.TestNamespace(client, "kube-system", test)
	k8s.TestNamespace(client, "local-path-storage", test)

	if platform.CertManager == nil || !platform.CertManager.Disabled {
		k8s.TestNamespace(client, "cert-manager", test)
	}

	if platform.Quack == nil || !platform.Quack.Disabled {
		k8s.TestNamespace(client, "quack", test)
	}

	if platform.Minio == nil || !platform.Minio.Disabled {
		k8s.TestNamespace(client, "minio", test)
	}

	if platform.GetIngressController() == "contour" {
		k8s.TestNamespace(client, "contour", test)
	} else {
		k8s.TestNamespace(client, "nginx-ingress", test)
	}

}
