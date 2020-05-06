package base

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	client, _ := platform.GetClientset()
	k8s.TestNamespace(client, "kube-system", test)
	k8s.TestNamespace(client, "local-path-storage", test)
	k8s.TestNamespace(client, "cert-manager", test)

	if platform.Nginx == nil || !platform.Nginx.Disabled {
		k8s.TestNamespace(client, "ingress-nginx", test)
	}

	if platform.Minio == nil || !platform.Minio.Disabled {
		k8s.TestNamespace(client, "minio", test)
	}

	if platform.E2E {
		TestPlatformOperatorAutoDeleteNamespace(platform, test)
		TestPlatformOperatorPodAnnotations(platform, test)
	}
}

func TestPlatformOperatorAutoDeleteNamespace(p *platform.Platform, test *console.TestResults) {
	namespace := fmt.Sprintf("platform-operator-e2e-auto-delete-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	annotations := map[string]string{
		"auto-delete": "10s",
	}

	if err := p.CreateOrUpdateNamespace(namespace, nil, annotations); err != nil {
		test.Failf("platform-operator", "failed to create namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(namespace, nil) // nolint: errcheck
	}()

	time.Sleep(15 * time.Second)

	ns, err := client.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		test.Failf("platform-operator", "failed to get namespace %s: %v", ns, err)
	}

	test.Passf("platform-operator", "Successfully cleaned up namespace %s with auto-delete=10s", namespace)

}

func TestPlatformOperatorPodAnnotations(p *platform.Platform, test *console.TestResults) {
	namespace := fmt.Sprintf("platform-operator-e2e-pod-annotations-%s", utils.RandomString(6))
	client, _ := p.GetClientset()

	annotationKey := "foo.flanksource.com/bar"
	annotationValue := utils.RandomString(6)
	annotationKey2 := "foo.flanksource.com/ignored"
	annotations := map[string]string{
		annotationKey:  annotationValue,
		annotationKey2: utils.RandomString(6),
	}

	if err := p.CreateOrUpdateNamespace(namespace, nil, annotations); err != nil {
		test.Failf("platform-operator", "failed to create namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(namespace, nil) // nolint: errcheck
	}()

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

	if _, err := client.CoreV1().Pods(namespace).Create(pod); err != nil {
		test.Failf("platform-operator", "failed to create pod %s in namespace %s: %v", podName, namespace, err)
		return
	}

	fetchedPod, err := client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		test.Failf("platform-operator", "failed to get pod %s in namespace %s: %v", podName, namespace, err)
		return
	}

	if fetchedPod.Annotations == nil {
		test.Failf("platform-operator", "failed to find any annotations for pod %s in namespace %s: %v", podName, namespace, err)
		return
	}
	if fetchedPod.Annotations[annotationKey] != annotationValue {
		test.Failf("platform-operator", "expected to have %s=%s got %s=%s", annotationKey, annotationValue, annotationKey, fetchedPod.Annotations[annotationKey])
		return
	}

	if fetchedPod.Annotations[annotationKey2] != "" {
		test.Failf("platform-operator", "expected key %s was not inherited from namespace %s", annotationKey2)
		return
	}

	test.Passf("platform-operator", "Pod %s inherits annotations from namespace", podName)
}
