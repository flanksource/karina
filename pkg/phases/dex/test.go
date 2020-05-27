package dex

import (
	"fmt"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	testlib "github.com/moshloop/platform-cli/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Import kubernetes oidc auth plugin
	"k8s.io/client-go/tools/clientcmd"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Ldap.Disabled {
		test.Skipf("dex", "LDAP is disabled in platform config - skipping.")
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "dex", test)
	if !p.E2E {
		return
	}
	k8s.TestNamespace(client, "ldap", test)

	dexClient := &testlib.DexOauth{
		DexURL:       fmt.Sprintf("https://dex.%s", p.Domain),
		ClientID:     "kubernetes",
		ClientSecret: "ZXhhbXBsZS1hcHAtc2VjcmV0",
		RedirectURI:  "http://localhost:8000",
		Username:     p.Ldap.E2E.Username,
		Password:     p.Ldap.E2E.Password,
	}

	token, err := dexClient.GetAccessToken()
	if err != nil {
		test.Failf("dex", "failed to get token %v", err)
		return
	}

	test.Passf("dex", "OIDC Authentication flow")

	ca := p.GetIngressCA()
	kubeConfig, err := k8s.CreateOIDCKubeConfig(p.Name, ca, "localhost", fmt.Sprintf("https://dex.%s", p.Domain), token.IDToken, token.AccessToken, token.RefreshToken)

	if err != nil {
		test.Failf("dex", "failed to generate kube config: %v", err)
		return
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		test.Failf("dex", "failed to get rest client config: %v", err)
		return
	}

	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		test.Failf("dex", "failed to create k8s config: %v", err)
		return
	}
	pods, err := k8s.CoreV1().Pods("dex").List(metav1.ListOptions{})
	if err != nil {
		test.Failf("dex", "failed to list pods: %v", err)
		return
	}
	for _, pod := range pods.Items {
		test.Passf("dex", "%s => %s", pod.Name, pod.Status.Phase)
	}
	test.Passf("dex", "OIDC Kubeconfig access")
}
