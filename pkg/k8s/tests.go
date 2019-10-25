package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/moshloop/commons/console"
)

func TestNamespace(client kubernetes.Interface, ns string, t *console.TestResults) {
	list, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Failf("Failed to get pods for %s: %v", ns, err)
		return
	}

	for _, pod := range list.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
			t.Passf(ns, "%s => %s", pod.Name, pod.Status.Phase)
		} else {
			t.Failf(ns, "%s => %s ", pod.Name, pod.Status.Phase)
		}
	}
	// check all pods running or completed with < 3 restarts
	// check unbound pvcs
	// check all pod liveness / readiness
}
