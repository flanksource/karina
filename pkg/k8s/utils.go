package k8s

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetValidName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

func WaitForNamespace(client kubernetes.Interface, ns string, timeout time.Duration) {
	pods := client.CoreV1().Pods(ns)
	start := time.Now()
	for {
		ready := 0
		pending := 0
		list, _ := pods.List(metav1.ListOptions{})
		for _, pod := range list.Items {
			conditions := true
			for _, condition := range pod.Status.Conditions {
				if condition.Status == v1.ConditionFalse {
					conditions = false
				}
			}
			if conditions && (pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded) {
				ready++
			} else {
				pending++
			}
		}
		if ready > 0 && pending == 0 {
			return
		}
		log.Debugf("ns/%s: ready=%d, pending=%d", ns, ready, pending)
		if start.Add(timeout).Before(time.Now()) {
			log.Warnf("ns/%s: ready=%d, pending=%d", ns, ready, pending)
			return
		}
		time.Sleep(10 * time.Second)
	}
}

func NewDeployment(ns, name, image string, labels map[string]string, port int32, args ...string) *apps.Deployment {
	if labels == nil {
		labels = make(map[string]string)
	}
	if len(labels) == 0 {
		labels["app"] = name
	}
	replicas := int32(1)

	deployment := apps.Deployment{
		ObjectMeta: NewObjectMeta(ns, name),
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: "IfNotPresent",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: port,
								},
							},
							Args:      args,
							Resources: LowResourceRequirements(),
						},
					},
				},
			},
		},
	}
	deployment.Kind = "Deployment"
	deployment.APIVersion = "apps/v1"
	return &deployment
}

func NewObjectMeta(ns, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
}

func LowResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("512Mi"),
			v1.ResourceCPU:    resource.MustParse("500m"),
		},
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("128Mi"),
			v1.ResourceCPU:    resource.MustParse("10m"),
		},
	}
}

func decodeStringToTimeDuration(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(time.Duration(5)) {
		return data, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToTimeDuration: Failed to parse duration: %v", err)
	}
	return d, nil
}

func decodeStringToDuration(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Duration{Duration: time.Duration(5)}) {
		return data, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToDuration: Failed to parse duration: %v", err)
	}
	return metav1.Duration{Duration: d}, nil
}

func decodeStringToTime(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Time{Time: time.Now()}) {
		return data, nil
	}
	d, err := time.Parse(time.RFC3339, data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToTime: failed to decode to time: %v", err)
	}
	return metav1.Time{Time: d}, nil
}

func IsPodCrashLoopBackoff(pod v1.Pod) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason == "CrashLoopBackOff" {
			return true
		}
	}
	return false
}

func IsPodHealthy(pod v1.Pod) bool {
	conditions := true
	for _, condition := range pod.Status.Conditions {
		if condition.Status == v1.ConditionFalse {
			conditions = false
		}
	}
	return conditions && (pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded)
}

func IsPodFinished(pod v1.Pod) bool {
	return pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed
}

func IsPodPending(pod v1.Pod) bool {
	return pod.Status.Phase == v1.PodPending
}

func IsMasterNode(node v1.Node) bool {
	_, ok := node.Labels["node-role.kubernetes.io/master"]
	return ok
}

func IsDeleted(object metav1.Object) bool {
	return object.GetDeletionTimestamp() != nil && !object.GetDeletionTimestamp().IsZero()
}

func IsPodDaemonSet(pod v1.Pod) bool {
	controllerRef := metav1.GetControllerOf(&pod)
	return controllerRef != nil && controllerRef.Kind == apps.SchemeGroupVersion.WithKind("DaemonSet").Kind
}

func GetNodeStatus(node v1.Node) string {
	s := ""
	for _, condition := range node.Status.Conditions {
		if condition.Status == v1.ConditionFalse {
			continue
		}
		if s != "" {
			s += ", "
		}
		s += string(condition.Type)
	}
	return s
}

type Health struct {
	RunningPods, PendingPods, ErrorPods, CrashLoopBackOff int
	ReadyNodes, UnreadyNodes                              int
	Error                                                 error
}

func (h Health) IsDegradedComparedTo(h2 Health) bool {
	if h2.RunningPods > h.RunningPods ||
		h.PendingPods-1 > h2.PendingPods || h.ErrorPods > h2.ErrorPods || h.CrashLoopBackOff > h2.CrashLoopBackOff {
		return true
	}
	return h.UnreadyNodes > h2.UnreadyNodes
}

func (h Health) String() string {
	return fmt.Sprintf("pods(running=%d, pending=%d, crashloop=%d, error=%d)  nodes(ready=%d, notready=%d)",
		h.RunningPods, h.PendingPods, h.CrashLoopBackOff, h.ErrorPods, h.ReadyNodes, h.UnreadyNodes)
}
