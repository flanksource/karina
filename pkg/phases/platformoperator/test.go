package platformoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"

	platformv1 "github.com/flanksource/karina/pkg/api/platformoperator/v1"
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
		TestPlatformOperatorClusterResourceQuota(platform, test)
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

func TestPlatformOperatorClusterResourceQuota(p *platform.Platform, test *console.TestResults) {
	if p.PlatformOperator.IsDisabled() {
		return
	}
	matchBy := map[string]string{
		"owner": fmt.Sprintf("group-%s", utils.RandomString(6)),
	}
	testName := "cluster-resource-quota"
	namespace1 := fmt.Sprintf("platform-operator-e2e-resource-quota1-%s", utils.RandomString(6))
	namespace2 := fmt.Sprintf("platform-operator-e2e-resource-quota2-%s", utils.RandomString(6))
	crqName := fmt.Sprintf("e2e-cluster-resource-quota-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	if err := p.CreateOrUpdateWorkloadNamespace(namespace1, matchBy, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace1, err)
		return
	}

	if err := p.CreateOrUpdateWorkloadNamespace(namespace2, matchBy, nil); err != nil {
		test.Failf(testName, "failed to create namespace %s: %v", namespace2, err)
		return
	}

	defer func() {
		if p.PlatformConfig.Trace {
			return
		}
		_ = client.CoreV1().Namespaces().Delete(context.TODO(), namespace1, metav1.DeleteOptions{})
		_ = client.CoreV1().Namespaces().Delete(context.TODO(), namespace2, metav1.DeleteOptions{})
		_ = p.DeleteByKind("ClusterResourceQuota", "", crqName)
	}()

	if err := p.Apply("", newClusterResourceQuota(crqName, "5", "8Gi", matchBy)); err != nil {
		test.Failf(testName, "Failed to create ClusterResourceQuota: %v", err)
		return
	}

	if err := p.Apply("", newResourceQuota("resource-quota1", namespace1, "2", "4Gi")); err != nil {
		test.Failf(testName, "expected to create ResourceQuota %v", err)
		return
	}

	if err := p.Apply("", newResourceQuota("resource-quota2", namespace2, "2", "1Gi")); err != nil {
		test.Failf(testName, "expected to create ResourceQuota %v", err)
		return
	}

	if err := p.Apply("", newResourceQuota("resource-quota3", namespace1, "2", "6Gi")); err == nil {
		test.Failf(testName, "Expected resource-quota above cluster resource quota to be denied")
		return
	}
	test.Passf(testName, "cluster resource quotas work as expected")
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

func newClusterResourceQuota(name, cpu, memory string, matchBy map[string]string) *platformv1.ClusterResourceQuota {
	crq := &platformv1.ClusterResourceQuota{
		TypeMeta:   metav1.TypeMeta{APIVersion: "platform.flanksource.com/v1", Kind: "ClusterResourceQuota"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: platformv1.ClusterResourceQuotaSpec{
			MatchLabels: matchBy,
			ResourceQuotaSpec: v1.ResourceQuotaSpec{
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
