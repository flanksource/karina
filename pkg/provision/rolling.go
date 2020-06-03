package provision

import (
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/commons/timer"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RollingOptions struct {
	Timeout                time.Duration
	MinAge                 time.Duration
	Max                    int
	Force                  bool
	ScaleSingleDeployments bool
	MigrateLocalVolumes    bool
	Masters, Workers       bool
}

// Perform a rolling update of nodes
func RollingUpdate(platform *platform.Platform, opts RollingOptions) error {
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}
	rolled := 0
	for _, nodeMachine := range cluster.Nodes {
		machine := nodeMachine.Machine
		node := nodeMachine.Node
		age := machine.GetAge()
		template := machine.GetTemplate()

		if k8s.IsMasterNode(node) && !opts.Masters {
			continue
		} else if !k8s.IsMasterNode(node) && !opts.Workers {
			continue
		}
		if age > opts.MinAge {
			platform.Infof("Replacing %s,  age=%s, template=%s ", machine.Name(), age, template)
		} else {
			continue
		}

		if err := cluster.Cordon(node); err != nil {
			return err
		}

		health := platform.GetHealth()

		platform.Infof("Health Before: %s", health)

		timer := timer.NewTimer()

		var replacement types.Machine
		if k8s.IsMasterNode(node) {
			replacement, err = createSecondaryMaster(platform)
			if err != nil {
				return fmt.Errorf("failed to create new secondary master: %v", err)
			}
		} else {
			replacement, err = createWorker(platform, "")
			if err != nil {
				return fmt.Errorf("failed to create new worker: %v", err)
			}
		}

		platform.Infof("[%s] waiting for replacement to become ready", replacement.Name())
		if status, err := platform.WaitForNode(replacement.Name(), opts.Timeout, v1.NodeReady, v1.ConditionTrue); err != nil {
			return fmt.Errorf("[%s] replacement did not come up healthy: %v", replacement.Name(), status)
		}

		terminate(platform, machine)

		platform.Infof("Replaced %s in %s", node.Name, timer)
		rolled++
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
			return fmt.Errorf("cluster is not healthy, aborting rollout after %d of %d ", rolled, cluster.Nodes.Len())
		}
		if rolled >= opts.Max {
			break
		}
	}
	platform.Infof("Rollout finished, rolled %d of %d ", rolled, cluster.Nodes.Len())
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

		timer := timer.NewTimer()
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
