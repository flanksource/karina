package k8s

import (
	"fmt"

	"github.com/flanksource/commons/console"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestNamespace(client kubernetes.Interface, ns string, t *console.TestResults) {
	pods := client.CoreV1().Pods(ns)
	events := client.CoreV1().Events(ns)
	list, err := pods.List(metav1.ListOptions{})
	if err != nil {
		t.Failf("Failed to get pods for %s: %v", ns, err)
		return
	}

	if len(list.Items) == 0 {
		_, err := client.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			t.Skipf(ns, "[%s] namespace not found, skipping", ns)
		} else {
			t.Failf(ns, "[%s] Expected pods but none running - did you deploy?", ns)
		}
	}
	for _, pod := range list.Items {
		conditions := true
		// for _, condition := range pod.Status.Conditions {
		// 	if condition.Status == v1.ConditionFalse {
		// 		t.Failf(ns, "%s => %s: %s", pod.Name, condition.Type, condition.Message)
		// 		conditions = false
		// 	}
		// }
		if conditions && pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
			t.Passf(ns, "%s => %s", pod.Name, pod.Status.Phase)
		} else {
			events, err := events.List(metav1.ListOptions{
				FieldSelector: "involvedObject.name=" + pod.Name,
			})
			if err != nil {
				t.Failf(ns, "%s => %s, failed to get events %+v ", pod.Name, pod.Status.Phase, err)
				continue
			}
			msg := ""
			for _, event := range events.Items {
				if event.Type == "Normal" {
					continue
				}
				msg += fmt.Sprintf("%s: %s ", event.Reason, event.Message)
			}
			t.Failf(ns, "%s/%s=%s %s ", ns, pod.Name, pod.Status.Phase, msg)
		}
	}
	// check all pods running or completed with < 3 restarts
	// check unbound pvcs
	// check all pod liveness / readiness
}
