package nodelocaldns

import (
	"context"
	"fmt"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const Namespace = constants.KubeSystem

func Install(platform *platform.Platform) error {
	if platform.NodeLocalDNS.Disabled {
		return platform.DeleteSpecs(Namespace, "node-local-dns.yaml")
	}
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}

	kubeDNS, err := client.CoreV1().Services("kube-system").Get(context.TODO(), "kube-dns", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("install: Failed to get service: %v", err)
	}

	platform.NodeLocalDNS.DNSServer = kubeDNS.Spec.ClusterIP

	platform.NodeLocalDNS.LocalDNS = "169.254.20.10"
	platform.NodeLocalDNS.DNSDomain = "cluster.local"

	return platform.ApplySpecs(Namespace, "node-local-dns.yaml")
}
