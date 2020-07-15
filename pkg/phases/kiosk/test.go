package kiosk

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	kioskapi "github.com/flanksource/karina/pkg/api/kiosk/tenancy/v1alpha1"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Kiosk.IsDisabled() {
		return
	}

	client, _ := p.GetClientset()
	k8s.TestDeploy(client, constants.PlatformSystem, "kiosk", test)

	if !p.E2E {
		return
	}

	user1Client, err := getImpersonateClient(p, "user1")
	if err != nil {
		test.Failf("kiosk", "failed to get impersonate client for user user1: %v", err)
		return
	}

	TestUserDirectNamespaceAccess(p, test, user1Client)
	TestUserCreateSpace(p, test, user1Client)
}

func TestUserDirectNamespaceAccess(p *platform.Platform, test *console.TestResults, user1Client *kubernetes.Clientset) {
	_, err := user1Client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err == nil {
		test.Failf("kiosk", "expected user1 to not be able to list namespaces")
	} else if err.Error() != "namespaces is forbidden: User \"user1\" cannot list resource \"namespaces\" in API group \"\" at the cluster scope" {
		test.Failf("kiosk", "received unexpected error: %v", err)
	} else {
		test.Passf("kiosk", "user1 is not able to list namespaces through the API")
	}

	name := fmt.Sprintf("test-namespace-%s", utils.RandomString(6))
	ns := &v1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	_, err = user1Client.CoreV1().Namespaces().Create(ns)
	if err == nil {
		test.Failf("kiosk", "expected user1 to not be able to create namespaces")
	} else if err.Error() != "namespaces is forbidden: User \"user1\" cannot create resource \"namespaces\" in API group \"\" at the cluster scope" {
		test.Failf("kiosk", "received unexpected error: %v", err)
	} else {
		test.Passf("kiosk", "user1 is not able to create namespaces through the API")
	}
}

func TestUserCreateSpace(p *platform.Platform, test *console.TestResults, user1CLient *kubernetes.Clientset) {
	// Create Account for user1

	space := &kioskapi.Space{
		TypeMeta: metav1.TypeMeta{Kind: "Space", APIVersion: "tenancy.kiosk.sh/v1alpha1"},
	}
	spaceClient, _, _, err := p.GetDynamicClientFor("", space)
	if err != nil {
		test.Failf("kiosk", "failed to get dynamic client: %v", err)
		return
	}

	spaces, err := spaceClient.List(metav1.ListOptions{})
	if err != nil {
		test.Failf("kiosk", "failed to list spaces: %v", err)
		return
	}

	for _, s := range spaces.Items {
		p.Infof("Found space: %s", s.GetName())
	}
}

func getImpersonateClient(p *platform.Platform, username string) (*kubernetes.Clientset, error) {
	data, err := p.GetKubeConfigBytes()
	if err != nil {
		return nil, fmt.Errorf("getRESTConfig: failed to get kubeconfig: %v", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("kubeConfig is empty")
	}

	rc, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rest config from kube config")
	}

	impersonate := transport.ImpersonationConfig{UserName: "user1"}

	transportConfig, err := rc.TransportConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transport config")
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tls config")
	}
	timeout := 5 * time.Second

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
			DualStack: false, // K8s do not work well with IPv6
		}).DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: 10 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       20 * time.Second,
		TLSClientConfig:       tlsConfig,
	}

	rc.Transport = transport.NewImpersonatingRoundTripper(impersonate, tr)
	rc.TLSClientConfig = rest.TLSClientConfig{}
	user1Client, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kubernetes config")
	}

	return user1Client, nil
}
