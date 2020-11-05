package burnin

import (
	"time"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	Taint = "node.kubernetes.io/burnin"
)

func Run(platform *platform.Platform, period time.Duration, quit chan bool) {
	for {
		if err := reconcile(platform, period); err != nil {
			platform.Errorf("Error reconciling: %v", err)
		}
		select {
		case <-quit:
			return
		default:
			time.Sleep(20 * time.Second)
		}
	}
}

func reconcile(platform *platform.Platform, period time.Duration) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

nodeLoop:
	for _, node := range nodes.Items {
		if !kommons.HasTaint(node, Taint) {
			continue
		}
		podList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node.Name}).String()})
		if err != nil {
			return err
		}

		for _, pod := range podList.Items {
			if !kommons.IsPodHealthy(pod) {
				platform.Infof("Node is not healthy yet, pod is unhealthy pod=%s node=%s", pod.Name, node.Name)
				continue nodeLoop
			}
			lastRestartTime := kommons.GetLastRestartTime(pod)
			if lastRestartTime != nil && time.Since(*lastRestartTime) < period {
				platform.Infof("Node is not healthy yet, pod restarted %s ago pod=%s node=%s", time.Since(*lastRestartTime), pod.Name, node.Name)
				continue nodeLoop
			}
			if time.Since(pod.CreationTimestamp.Time) < period {
				platform.Infof("Node is not healthy yet, pod is too young, create %s ago pod=%s node=%s", time.Since(pod.CreationTimestamp.Time), pod.Name, node.Name)
				continue nodeLoop
			}
		}
		// everything looks healthy lets remove the taint
		platform.Infof("Removing burnin taint node=%s", node.Name)
		node.Spec.Taints = kommons.RemoveTaint(node.Spec.Taints, Taint)
		if _, err := client.CoreV1().Nodes().Update(&node); err != nil {
			return err
		}
	}
	return nil
}
