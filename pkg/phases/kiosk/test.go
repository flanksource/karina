package kiosk

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	kioskconfigapi "github.com/flanksource/karina/pkg/api/kiosk/config/v1alpha1"
	kioskapi "github.com/flanksource/karina/pkg/api/kiosk/tenancy/v1alpha1"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
	"k8s.io/utils/pointer"
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

	TestUserDirectNamespaceAccess(p, test)
	TestUserCreateSpace(p, test)
}

func TestUserDirectNamespaceAccess(p *platform.Platform, test *console.TestResults) {
	user1Client, err := getImpersonateClient(p, "user1")
	if err != nil {
		test.Failf("kiosk", "failed to get impersonate client for user user1: %v", err)
		return
	}

	_, err = user1Client.CoreV1().Namespaces().List(metav1.ListOptions{})
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

	// Test that user1 does not have access to list spaces until we create an Account

	space := &kioskapi.Space{
		TypeMeta:   metav1.TypeMeta{Kind: "Space", APIVersion: "tenancy.kiosk.sh/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-space-%s", utils.RandomString(6))},
	}
	spaceClient, _, unstructuredObject, err := p.GetDynamicClientForUser("", space, "user1")
	if err != nil {
		test.Failf("kiosk", "failed to get dynamic client: %v", err)
		return
	}

	_, err = spaceClient.List(metav1.ListOptions{})
	if err == nil {
		test.Failf("kiosk", "expected user1 to not be able to list spaces")
	} else if err.Error() != "spaces.tenancy.kiosk.sh is forbidden: User \"user1\" cannot list resource \"spaces\" in API group \"tenancy.kiosk.sh\" at the cluster scope" {
		test.Failf("kiosk", "received unexpected error: %v", err)
	} else {
		test.Passf("kiosk", "user1 is not able to list spaces through the API without Account")
	}

	_ = unstructuredObject
	// _, err = spaceClient.Create(unstructuredObject, metav1.CreateOptions{})
	// if err == nil {
	// 	test.Failf("kiosk", "expected user1 to not be able to create spaces")
	// } else if err.Error() != "spaces.tenancy.kiosk.sh is forbidden: User \"user1\" cannot create resource \"spaces\" in API group \"tenancy.kiosk.sh\" at the cluster scope" {
	// 	test.Failf("kiosk", "received unexpected error: %v", err)
	// } else {
	// 	test.Passf("kiosk", "user1 is not able to create spaces through the API without Account")
	// }
}

func TestUserCreateSpace(p *platform.Platform, test *console.TestResults) {
	key := utils.RandomString(6)
	user := fmt.Sprintf("user-%s", key)
	accountName := fmt.Sprintf("account-%s", key)

	// Create Account for user1
	account := &kioskapi.Account{
		TypeMeta:   metav1.TypeMeta{Kind: "Account", APIVersion: "tenancy.kiosk.sh/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: accountName},
		Spec: kioskapi.AccountSpec{
			AccountSpec: kioskconfigapi.AccountSpec{
				Subjects: []rbacv1.Subject{
					rbacv1.Subject{
						Kind:     "User",
						Name:     user,
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
				Space: kioskconfigapi.AccountSpace{
					ClusterRole: pointer.StringPtr("kiosk-space-admin"),
				},
			},
		},
	}

	accountClient, _, accountObj, err := p.GetDynamicClientFor("", account)
	if err != nil {
		test.Failf("kiosk", "failed to get dynamic client for accounts: %v", err)
		return
	}
	if _, err := accountClient.Create(accountObj, metav1.CreateOptions{}); err != nil {
		test.Failf("kiosk", "failed to create %s Account: %v", user, err)
		return
	}
	defer func() {
		if err := accountClient.Delete(account.Name, nil); err != nil {
			p.Errorf("failed to delete account %s", account.Name)
		}
	}()

	space := &kioskapi.Space{
		TypeMeta:   metav1.TypeMeta{Kind: "Space", APIVersion: "tenancy.kiosk.sh/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("user1-space-%s", utils.RandomString(6))},
		Spec:       kioskapi.SpaceSpec{Account: account.Name},
	}
	spaceClient, _, spaceObj, err := p.GetDynamicClientForUser("", space, user)
	if err != nil {
		test.Failf("kiosk", "failed to get dynamic client for spaces: %v", err)
		return
	}

	forbiddenStr := "spaces.tenancy.kiosk.sh is forbidden"
	err = nil
	var spaces *unstructured.UnstructuredList
	// Wait until Account has permissions to list spaces
	doUntil(func() bool {
		spaces, err = spaceClient.List(metav1.ListOptions{})
		if err != nil && !strings.HasPrefix(err.Error(), forbiddenStr) {
			test.Failf("kiosk", "failed to list spaces: %v", err)
			return true
		}
		err = nil

		return err != nil
	})

	if err != nil {
		test.Failf("kiosk", "received unexpected error: %v", err)
		return
	}

	if len(spaces.Items) != 0 {
		test.Failf("kiosk", "expected user %s to see no spaces after Account is created: %v", user, err)
		return
	}

	test.Passf("kiosk", "user %s can see 0 spaces after Account is created", user)

	if _, err := spaceClient.Create(spaceObj, metav1.CreateOptions{}); err != nil {
		test.Failf("kiosk", "failed to create space %s by user %s: %v", space.Name, user, err)
		return
	}

	test.Passf("kiosk", "user %s created space %s", user, space.Name)

	spaces, err = spaceClient.List(metav1.ListOptions{})
	if err != nil {
		test.Failf("kiosk", "failed to list spaces: %v", err)
		return
	}

	defer func() {
		adminSpaceClient, _, _, err := p.GetDynamicClientFor("", space)
		if err != nil {
			p.Errorf("failed to get admin space client: %v", err)
			return
		}
		if err := adminSpaceClient.Delete(space.Name, nil); err != nil {
			p.Errorf("failed to delete space %s: %v", space.Name, err)
		}
	}()

	if len(spaces.Items) != 1 {
		test.Failf("kiosk", "expected user %s to see 1 space, got %d", user, len(spaces.Items))
		return
	} else if spaces.Items[0].GetName() != space.Name {
		test.Failf("kiosk", "expected user %s to see space %s got %s", user, space.Name, spaces.Items[0].GetName())
		return
	} else {
		test.Passf("kiosk", "user %s can see 1 space: %s", user, space.Name)
	}

	k8s, err := getImpersonateClient(p, user)
	if err != nil {
		test.Failf("kiosk", "failed to get impersonate client for user user1: %v", err)
		return
	}

	_, err = k8s.CoreV1().Namespaces().Get(space.Name, metav1.GetOptions{})
	if err != nil {
		test.Failf("kiosk", "failed to get namespace %s: %v", space.Name, err)
		return
	}

	test.Passf("kiosk", "user %s can get namespace %s", user, space.Name)

	err = spaceClient.Delete(space.Name, nil)
	if err != nil {
		test.Failf("kiosk", "Expected user %s to be able to delete space %s", user, space.Name)
		return
	}

	test.Passf("kiosk", "user %s deleted space %s", user, space.Name)
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

	impersonate := transport.ImpersonationConfig{UserName: username}

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

func doUntil(fn func() bool) bool {
	start := time.Now()

	for {
		if fn() {
			return true
		}
		// if time.Now().After(start.Add(5 * time.Minute)) {
		if time.Now().After(start.Add(15 * time.Second)) {
			return false
		}
		time.Sleep(5 * time.Second)
	}
}
