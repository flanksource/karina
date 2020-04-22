package provision

import (
	"fmt"
	"sort"
	"time"

	"github.com/moshloop/platform-cli/pkg/phases/kubeadm"
	"github.com/moshloop/platform-cli/pkg/platform"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RollingOptions struct {
	Timeout                time.Duration
	MinAge                 time.Duration
	Force                  bool
	ScaleSingleDeployments bool
	MigrateLocalVolumes    bool
}

// Perform a rolling restart of nodes
func RollingUpdate(platform *platform.Platform, opts RollingOptions) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}
	if err := WithVmwareCluster(platform); err != nil {
		return err
	}

	// make sure admin kubeconfig is available
	platform.GetKubeConfig() // nolint: errcheck
	if platform.JoinEndpoint == "" {
		platform.JoinEndpoint = "localhost:8443"
	}

	// upload control plane certs first
	if _, err := kubeadm.UploadControlPlaneCerts(platform); err != nil {
		return err
	}
	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range list.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			platform.Infof("Skipping master %s", node.Name)
			continue
		}

		machine, err := platform.Cluster.GetMachine(node.Name)
		if err != nil {
			return err
		}

		attributes, err := machine.GetAttributes()
		if err != nil {
			return err
		}
		created, _ := time.Parse("02Jan06-15:04:05", attributes["CreatedDate"])
		template := attributes["Template"]
		age := time.Since(created)
		if age > opts.MinAge {
			platform.Infof("Replacing %s, created=%s, age=%s, template=%s ", machine.Name(), created, age, template)
		} else {
			platform.Infof("Skipping %s,  created=%s, age=%s, template=%s ", machine.Name(), created, age, template)
			continue
		}

		health := platform.GetHealth()

		platform.Infof("Health Before: %s", health)

		timer := NewTimer()

		if err := platform.Cordon(node.Name); err != nil {
			return fmt.Errorf("failed to cordon %s: %v", node.Name, err)
		}
		replacement, err := createWorker(platform, "")
		if err != nil {
			return fmt.Errorf("failed to create new worker: %v", err)
		}

		platform.Infof("waiting for replacement for %s to become ready", node.Name)
		if status, err := platform.WaitForNode(replacement.Name(), opts.Timeout, v1.NodeReady, v1.ConditionTrue); err != nil {
			return fmt.Errorf("new worker %s did not come up healthy: %v", node.Name, status)
		}

		terminate(platform, machine)

		platform.Infof("Replaced %s in %s", node.Name, timer)

		if succeededWithinTimeout := doUntil(opts.Timeout, func() bool {
			currentHealth := platform.GetHealth()
			platform.Infof(currentHealth.String())
			if currentHealth.IsDegradedComparedTo(health) {
				time.Sleep(5 * time.Second)
				return false
			}
			return true
		}); !succeededWithinTimeout {
			platform.Errorf("Health degraded after waiting %v", timer)
		}
		if platform.GetHealth().IsDegradedComparedTo(health) {
			return fmt.Errorf("cluster is not healthy, aborting rollout")
		}
	}
	return nil
}

// Perform a rolling restart of nodes
func RollingRestart(platform *platform.Platform, opts RollingOptions) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	var names sort.StringSlice
	nodes := make(map[string]v1.Node)
	for _, node := range list.Items {
		names = append(names, node.Name)
		nodes[node.Name] = node
	}
	sort.Sort(sort.Reverse(names))
	for _, name := range names {
		node := nodes[name]
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			platform.Infof("Skipping master %s", node.Name)
			continue
		}

		health := platform.GetHealth()

		platform.Infof("Health Before: %s", health)

		timer := NewTimer()
		if err := platform.Drain(node.Name, opts.Timeout); err != nil {
			if opts.Force {
				platform.Errorf("failed to drain %s, force restarting: %v", node.Name, err)
			} else {
				return fmt.Errorf("failed to drain %s: %v", node.Name, err)
			}
		}
		// we issue a shutdown command for 1m in the future to allow the command to complete successfully
		if _, err := platform.Executef(node.Name, opts.Timeout, "/sbin/shutdown -r +1"); err != nil {
			platform.Infof("Error restarting node: %s: %v", node.Name, err)
		}
		// then wait for the shutdown to actually occur and the node to become unhealthy
		platform.Infof("[%s] waiting for node to shutdown (become NotReady)", node.Name)
		if status, err := platform.WaitForNode(node.Name, opts.Timeout, v1.NodeReady, v1.ConditionFalse, v1.ConditionUnknown); err != nil || (status[v1.NodeReady] != v1.ConditionFalse && status[v1.NodeReady] != v1.ConditionUnknown) {
			if opts.Force {
				platform.Errorf("timed out did not detect node becoming unready %s :%v", node.Name, status)
			} else {
				return fmt.Errorf("failed to restart node %s: %v", node.Name, status)
			}
		} else {
			platform.Infof("Node is %v", status)
		}
		platform.Infof("[%s] waiting for node to finish restarting (become Ready)", node.Name)
		if status, err := platform.WaitForNode(node.Name, opts.Timeout, v1.NodeReady, v1.ConditionTrue); err != nil {
			return fmt.Errorf("%s did not come back up: %v", node.Name, status)
		}
		if err := platform.Uncordon(node.Name); err != nil {
			return fmt.Errorf("failed to uncordon %s: %v", node.Name, err)
		}
		platform.Infof("Restarted %s in %s", node.Name, timer)

		if succeededWithinTimeout := doUntil(opts.Timeout, func() bool {
			currentHealth := platform.GetHealth()
			platform.Infof(currentHealth.String())
			if currentHealth.IsDegradedComparedTo(health) {
				time.Sleep(5 * time.Second)
				return false
			}
			return true
		}); !succeededWithinTimeout {
			platform.Warnf("Current health not recovered after timeout %v", opts.Timeout)
		}
	}
	return nil
}

func doUntil(timeout time.Duration, fn func() bool) bool {
	start := time.Now()

	for {
		if fn() {
			return true
		}
		if time.Now().After(start.Add(timeout)) {
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
