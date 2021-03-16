package platformoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"

	platformv1 "github.com/flanksource/karina/pkg/api/platformoperator"
	"github.com/flanksource/karina/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	if platform.PlatformOperator.IsDisabled() {
		return
	}

	if err := platform.WaitForDeployment(Namespace, WebhookService, 2*time.Minute); err != nil {
		test.Failf("platform-operator", "platform-operator did not come up: %v", err)
		return
	}
	test.Passf("platform-operator", "platform-operator is ready")
	if !platform.E2E {
		return
	}
	TestPlatformOperatorAutoDeleteNamespace(platform, test)
	TestPlatformOperatorPodAnnotations(platform, test)
	if platform.PlatformOperator.EnableClusterResourceQuota {
		TestPlatformOperatorClusterResourceQuota1(platform, test)
		TestPlatformOperatorClusterResourceQuota2(platform, test)
	}
}

func TestPlatformOperatorAutoDeleteNamespace(p *platform.Platform, test *console.TestResults) {
	testName := "platform-auto-delete"
	namespace := fmt.Sprintf("platform-operator-e2e-auto-delete-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	annotations := map[string]string{
		"auto-delete": "10s",
	}

	if err := p.CreateOrUpdateWorkloadNamespace(namespace, nil, annotations); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	time.Sleep(15 * time.Second)

	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		test.Failf(testName, "failed to get namespace %s: %v", ns, err)
	}

	test.Passf(testName, "Successfully cleaned up namespace %s with auto-delete=10s", namespace)
}

func TestPlatformOperatorPodAnnotations(p *platform.Platform, test *console.TestResults) {
	testName := "pod-annotator"
	namespace := fmt.Sprintf("platform-operator-e2e-pod-annotations-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	annotationKey := "foo.flanksource.com/bar"
	annotationValue := utils.RandomString(6)
	annotationKey2 := "foo.flanksource.com/ignored"
	annotations := map[string]string{
		annotationKey:  annotationValue,
		annotationKey2: utils.RandomString(6),
	}

	if err := p.CreateOrUpdateWorkloadNamespace(namespace, nil, annotations); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace, err)
		return
	}

	if !p.PlatformConfig.Trace {
		defer func() {
			client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{}) // nolint: errcheck
		}()
	}

	podName := fmt.Sprintf("test-pod-annotations-%s", utils.RandomString(6))
	pod := &v1.Pod{
		TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: namespace},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}

	if _, err := client.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "failed to create pod %s in namespace %s: %v", podName, namespace, err)
		return
	}

	fetchedPod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		test.Failf(testName, "failed to get pod %s in namespace %s: %v", podName, namespace, err)
		return
	}

	if fetchedPod.Annotations == nil {
		test.Failf(testName, "failed to find any annotations for pod %s in namespace %s: %v", podName, namespace, err)
		return
	}
	if fetchedPod.Annotations[annotationKey] != annotationValue {
		test.Failf(testName, "expected to have %s=%s got %s=%s", annotationKey, annotationValue, annotationKey, fetchedPod.Annotations[annotationKey])
		return
	}

	if fetchedPod.Annotations[annotationKey2] != "" {
		test.Failf(testName, "expected key %s was not inherited from namespace", annotationKey2)
		return
	}

	test.Passf(testName, "Pod %s inherits annotations from namespace", podName)
}

func TestPlatformOperatorClusterResourceQuota1(p *platform.Platform, test *console.TestResults) {
	testName := "cluster-resource-quota"
	namespace1 := fmt.Sprintf("platform-operator-e2e-resource-quota1-%s", utils.RandomString(6))
	namespace2 := fmt.Sprintf("platform-operator-e2e-resource-quota2-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	if err := p.CreateOrUpdateWorkloadNamespace(namespace1, nil, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace1, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace1, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	if err := p.CreateOrUpdateWorkloadNamespace(namespace2, nil, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace2, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace2, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	crqName := fmt.Sprintf("e2e-cluster-resource-quota-%s", utils.RandomString(6))
	crq := &platformv1.ClusterResourceQuota{
		TypeMeta:   metav1.TypeMeta{APIVersion: "platform.flanksource.com/v1", Kind: "ClusterResourceQuota"},
		ObjectMeta: metav1.ObjectMeta{Name: crqName},
		Spec: platformv1.ClusterResourceQuotaSpec{
			Quota: v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("5"),
					v1.ResourceMemory: resource.MustParse("8Gi"),
				},
			},
		},
		Status: platformv1.ClusterResourceQuotaStatus{
			Namespaces: []platformv1.ResourceQuotaStatusByNamespace{},
		},
	}
	crqClient, _, unstructuredObj, err := p.GetDynamicClientFor("", crq)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}
	if _, err := crqClient.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "failed to create cluster resource quota: %v", err)
		return
	}
	defer removeClusterResourceQuota(p, crq, test)

	rqAPI1 := client.CoreV1().ResourceQuotas(namespace1)
	rqAPI2 := client.CoreV1().ResourceQuotas(namespace2)

	rq1 := newResourceQuota("resource-quota", namespace1, "2", "4Gi")
	if _, err := rqAPI1.Create(context.TODO(), rq1, metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "failed to create resource quota cpu=2 memory=4Gi: %v", err)
		return
	}

	time.Sleep(2 * time.Second)

	rq := newResourceQuota("resource-quota", namespace2, "4", "2Gi")
	_, err = rqAPI2.Create(context.TODO(), rq, metav1.CreateOptions{})
	if err == nil {
		removeResourceQuota(p, test, rq1, rq)
		test.Failf(testName, "expected to fail creating second resource quota with 4 cpu and 2Gi")
		return
	}
	test.Passf(testName, "resource quota with cpu=4 and memory=2Gi was not permitted as expected")

	rq = newResourceQuota("resource-quota", namespace2, "2", "7Gi")
	_, err = rqAPI2.Create(context.TODO(), rq, metav1.CreateOptions{})
	if err == nil {
		removeResourceQuota(p, test, rq1, rq)
		test.Failf(testName, "expected to fail creating second resource quota with cpu=2 and memory=7Gi")
		return
	}
	test.Passf(testName, "resource quota with cpu=2 and memory=7Gi was not permitted as expected")

	rq2 := newResourceQuota("resource-quota", namespace2, "2", "2Gi")
	_, err = rqAPI2.Create(context.TODO(), rq2, metav1.CreateOptions{})
	if err != nil {
		test.Failf(testName, "expected to create second resource quota with cpu=2 and memory=2Gi: %v", err)
		return
	}

	removeResourceQuota(p, test, rq1, rq2)
	removeClusterResourceQuota(p, crq, test)

	// Wait until all ClusterResourceQuota and ResourceQuota are removed
	// Otherwise the next test will not pass
	doUntil(func() bool {
		if _, err := rqAPI1.Get(context.TODO(), rq1.Name, metav1.GetOptions{}); !errors.IsNotFound(err) {
			return false
		}
		if _, err := rqAPI2.Get(context.TODO(), rq2.Name, metav1.GetOptions{}); !errors.IsNotFound(err) {
			return false
		}
		if _, err := crqClient.Get(context.TODO(), crq.Name, metav1.GetOptions{}); !errors.IsNotFound(err) {
			return false
		}
		return true
	})

	test.Passf(testName, "cluster resource quota test 1 passed")
}

func TestPlatformOperatorClusterResourceQuota2(p *platform.Platform, test *console.TestResults) {
	if p.PlatformOperator.IsDisabled() {
		return
	}
	testName := "cluster-resource-quota"
	namespace1 := fmt.Sprintf("platform-operator-e2e-resource-quota1-%s", utils.RandomString(6))
	namespace2 := fmt.Sprintf("platform-operator-e2e-resource-quota2-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	if err := p.CreateOrUpdateWorkloadNamespace(namespace1, nil, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace1, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace1, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	if err := p.CreateOrUpdateWorkloadNamespace(namespace2, nil, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace2, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace2, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	rqAPI1 := client.CoreV1().ResourceQuotas(namespace1)
	rqAPI2 := client.CoreV1().ResourceQuotas(namespace2)

	if _, err := rqAPI1.Create(context.TODO(), newResourceQuota("resource-quota", namespace1, "2", "4Gi"), metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "failed to create resource quota 1: %v", err)
		return
	}

	if _, err := rqAPI2.Create(context.TODO(), newResourceQuota("resource-quota", namespace2, "2", "6Gi"), metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "failed to create resource quota 2: %v", err)
		return
	}

	time.Sleep(2 * time.Second)

	crqName := fmt.Sprintf("e2e-cluster-resource-quota-%s", utils.RandomString(6))
	crq := newClusterResourceQuota(crqName, "5", "8Gi")
	crqClient, _, unstructuredObj, err := p.GetDynamicClientFor("", crq)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}
	if _, err := crqClient.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{}); err == nil {
		removeClusterResourceQuota(p, crq, test)
		test.Failf(testName, "expected to fail creating cluster resource quota with cpu=5 and memory=8Gi")
		return
	}
	test.Passf(testName, "cluster resource quota with cpu=5 and memory=8Gi failed to create as expected")

	crq = newClusterResourceQuota(crqName, "3", "12Gi")
	crqClient, _, unstructuredObj, err = p.GetDynamicClientFor("", crq)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}
	if _, err := crqClient.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{}); err == nil {
		removeClusterResourceQuota(p, crq, test)
		test.Failf(testName, "expected to fail creating cluster resource quota with cpu=3 and memory=12Gi")
		return
	}
	test.Passf(testName, "cluster resource quota with cpu=3 and memory=12Gi failed to create as expected")

	crq = newClusterResourceQuota(crqName, "5", "12Gi")
	crqClient, _, unstructuredObj, err = p.GetDynamicClientFor("", crq)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}
	if _, err := crqClient.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{}); err != nil {
		test.Failf(testName, "expected to create cluster resource quota with 5 cpu and 12 Gi: %v", err)
		return
	}
	removeClusterResourceQuota(p, crq, test)
}

// nolint: unparam
func newResourceQuota(name, namespace, cpu, memory string) *v1.ResourceQuota {
	rq := &v1.ResourceQuota{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ResourceQuota"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(cpu),
				v1.ResourceMemory: resource.MustParse(memory),
			},
		},
	}
	return rq
}

func newClusterResourceQuota(name, cpu, memory string) *platformv1.ClusterResourceQuota {
	crq := &platformv1.ClusterResourceQuota{
		TypeMeta:   metav1.TypeMeta{APIVersion: "platform.flanksource.com/v1", Kind: "ClusterResourceQuota"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: platformv1.ClusterResourceQuotaSpec{
			Quota: v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
			},
		},
		Status: platformv1.ClusterResourceQuotaStatus{
			Namespaces: []platformv1.ResourceQuotaStatusByNamespace{},
		},
	}
	return crq
}

func removeClusterResourceQuota(p *platform.Platform, crq *platformv1.ClusterResourceQuota, test *console.TestResults) {
	apiClient, _, _, err := p.GetDynamicClientFor("", crq)
	if err != nil {
		test.Errorf("Failed to get dynamic client: %v", err)
		return
	}

	if err := apiClient.Delete(context.TODO(), crq.Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		test.Warnf("Failed to delete cluster resource quota %s: %v", crq.Name, err)
		return
	}
}

func removeResourceQuota(p *platform.Platform, test *console.TestResults, rqs ...*v1.ResourceQuota) {
	client, _ := p.GetClientset()
	for _, rq := range rqs {
		if err := client.CoreV1().ResourceQuotas(rq.Namespace).Delete(context.TODO(), rq.Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			test.Errorf("failed to delete resource quota %s in namespace %s: %v", rq.Name, rq.Namespace, err)
		}
	}
}

func doUntil(fn func() bool) bool {
	start := time.Now()

	for {
		if fn() {
			return true
		}
		if time.Now().After(start.Add(5 * time.Minute)) {
			return false
		}
		time.Sleep(5 * time.Second)
	}
}
