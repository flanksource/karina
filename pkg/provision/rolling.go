package provision

import (
	"fmt"
	"time"

	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Perform a rolling restart of nodes
func RollingUpdate(platform *platform.Platform, drainTimeout time.Duration, forceRestart bool) error {
	return nil
}

// Perform a rolling restart of nodes
func RollingRestart(platform *platform.Platform, drainTimeout time.Duration, forceRestart bool) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range list.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			log.Infof("Skipping master %s", node.Name)
			continue
		}

		health := platform.GetHealth()

		log.Infof("Health Before: %s", health)

		timer := NewTimer()
		if err := platform.Drain(node.Name, drainTimeout); err != nil {
			if forceRestart {
				log.Errorf("failed to drain %s, force restarting: %v", node.Name, err)
			} else {
				return fmt.Errorf("failed to drain %s: %v", node.Name, err)
			}
		}
		// we issue a shutdown command for 1m in the future to allow the command to complete successfully
		if _, err := platform.Executef(node.Name, drainTimeout, "/sbin/shutdown -r +1"); err != nil {
			log.Infof("Error restarting node: %s: %v", node.Name, err)
		}
		// then wait for the shutdown to actually occur and the node to become unhealthy
		log.Infof("[%s] waiting for node to shutdown (become NotReady)", node.Name)
		if status, err := platform.WaitForNode(node.Name, drainTimeout, v1.NodeReady, v1.ConditionFalse, v1.ConditionUnknown); err != nil || (status[v1.NodeReady] != v1.ConditionFalse && status[v1.NodeReady] != v1.ConditionUnknown) {
			if forceRestart {
				log.Errorf("timed out did not detect node becoming unready %s :%v", node.Name, status)
			} else {
				return fmt.Errorf("failed to restart node %s: %v", node.Name, status)
			}
		} else {
			log.Infof("Node is %v", status)
		}
		log.Infof("[%s] waiting for node to finish restarting (become Ready)", node.Name)
		if status, err := platform.WaitForNode(node.Name, drainTimeout, v1.NodeReady, v1.ConditionTrue); err != nil {
			return fmt.Errorf("%s did not come back up: %v", node.Name, status)
		}
		if err := platform.Uncordon(node.Name); err != nil {
			return fmt.Errorf("failed to uncordon %s: %v", node.Name, err)
		}
		log.Infof("Restarted %s in %s", node.Name, timer)

		doUntil(func() bool {
			currentHealth := platform.GetHealth()
			log.Infof(currentHealth.String())
			if currentHealth.IsDegradedComparedTo(health) {
				time.Sleep(5 * time.Second)
				return false
			}
			return true
		})
	}
	return nil
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

type Timer struct {
	Start time.Time
}

func (t Timer) Elapsed() float64 {
	return float64(time.Since(t.Start).Milliseconds())
}

func (t Timer) Millis() int64 {
	return time.Since(t.Start).Milliseconds()
}

func (t Timer) String() string {
	return fmt.Sprintf("%dms", t.Millis())
}

func NewTimer() Timer {
	return Timer{Start: time.Now()}
}
