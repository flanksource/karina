package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/moshloop/commons/console"
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
		t.Failf(ns, "Expected pods but none running - did you deploy?")
	}
	for _, pod := range list.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
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
			t.Failf(ns, "%s=%s %s ", pod.Name, pod.Status.Phase, msg)
		}
	}
	// check all pods running or completed with < 3 restarts
	// check unbound pvcs
	// check all pod liveness / readiness
}
