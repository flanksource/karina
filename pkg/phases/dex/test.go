package dex

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	testlib "github.com/moshloop/platform-cli/pkg/test"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "dex", test)
	k8s.TestNamespace(client, "ldap", test)

	dexClient := &testlib.DexOauth{
		DexURL:       fmt.Sprintf("https://dex.%s", p.Domain),
		ClientID:     "kubernetes",
		ClientSecret: "ZXhhbXBsZS1hcHAtc2VjcmV0",
		RedirectURI:  "http://localhost:8000",
		Username:     p.Ldap.Username,
		Password:     p.Ldap.Password,
	}

	token, err := dexClient.GetIDToken()
	if err != nil {
		test.Failf("dex", "failed to get token %v", err)
	}

	kubeConfig, err := testlib.GenerateKubeConfigOidc(p, token)
	if err != nil {
		test.Failf("dex", "failed to generate kube config: %v", err)
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		test.Failf("dex", "failed to get rest client config: %v", err)
	}

	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		test.Failf("dex", "failed to create k8s config: %v", err)
	}
	pods, err := k8s.CoreV1().Pods("dex").List(metav1.ListOptions{})
	if err != nil {
		test.Failf("dex", "failed to list pods: %v", err)
	}
	for _, pod := range pods.Items {
		test.Passf("dex", "%s => %s", pod.Name, pod.Status.Phase)
	}
}
