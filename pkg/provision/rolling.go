package provision

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
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
	MaxSurge               int
	HealthTolerance        int
	Force                  bool
	ScaleSingleDeployments bool
	MigrateLocalVolumes    bool
	Masters, Workers       bool
}

func replace(platform *platform.Platform, opts RollingOptions, cluster *Cluster, machine NodeMachine) error {
	node := machine.Node
	// first we cordon
	if err := cluster.Cordon(node); err != nil {
		return err
	}
	// then we surge up
	var replacement types.Machine
	var err error
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

	platform.Infof("[%s] waiting for replacement to become ready", replacement)
	if status, err := platform.WaitForNode(replacement.Name(), opts.Timeout, v1.NodeReady, v1.ConditionTrue); err != nil {
		return fmt.Errorf("[%s] replacement did not come up healthy: %v", replacement, status)
	}
	return nil
}

func selectMachinesToReplace(platform *platform.Platform, opts RollingOptions, cluster *Cluster) *NodeMachineBatch {
	toReplace := NodeMachineBatch{}
	// first we select all the nodes for replacement upfront
	for _, nodeMachine := range cluster.Nodes {
		machine := nodeMachine.Machine
		node := nodeMachine.Node
		age := machine.GetAge()
		template := machine.GetTemplate()

		// check if a node is available for update
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
		toReplace.Push(nodeMachine)
	}
	return &toReplace
}

type NodeMachineBatch []NodeMachine

func (h NodeMachineBatch) Len() int { return len(h) }
func (h NodeMachineBatch) Less(i, j int) bool {
	return h[i].Machine.GetAge() > h[j].Machine.GetAge()
}
func (h NodeMachineBatch) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *NodeMachineBatch) Push(x NodeMachine) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x)
}

func (h *NodeMachineBatch) Pop() NodeMachine {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *NodeMachineBatch) PopN(count int) *[]NodeMachine {
	items := []NodeMachine{}

	for i := 0; i < count; {
		if h.Len() == 0 || len(items) == count {
			return &items
		}
		items = append(items, h.Pop())
	}
	return &items
}

// Perform a rolling update of nodes
func RollingUpdate(platform *platform.Platform, opts RollingOptions) error {
	// collect all info about cluster and vm's
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}
	rolled := 0
	toReplace := selectMachinesToReplace(platform, opts, cluster)
	batch := toReplace.PopN(opts.MaxSurge)
	for len(*batch) > 0 {
		platform.Infof("Replacing %s, %d remaining", *batch, toReplace.Len())
		time.Sleep(10 * time.Second)
		timer := timer.NewTimer()
		health := platform.GetHealth()
		platform.Infof("Health Before: %s", health)
		batch = toReplace.PopN(opts.MaxSurge)
		wg := sync.WaitGroup{}
		var replacementError atomic.Value
		for _, nodeMachine := range *batch {
			_nodeMachine := nodeMachine
			wg.Add(1)
			go func() {
				if err := replace(platform, opts, cluster, _nodeMachine); err != nil {
					platform.Errorf(err.Error())
					replacementError.Store(err)
				}
				wg.Done()
			}()
			// force each replacement to have a different timestamp
			time.Sleep(2 * time.Second)
		}
		wg.Wait()
		if replacementError.Load() != nil {
			return replacementError.Load().(error)
		}

		// then we terminate the machines waiting to be replaced
		for _, nodeMachine := range *batch {
			terminate(platform, nodeMachine.Machine)
			rolled++
		}

		// finally we wait until we are the same health level as we were before
		if succeededWithinTimeout := doUntil(opts.Timeout, func() bool {
			currentHealth := platform.GetHealth()
			platform.Infof(currentHealth.String())
			if currentHealth.IsDegradedComparedTo(health, opts.HealthTolerance) {
				time.Sleep(5 * time.Second)
				return false
			}
			return true
		}); !succeededWithinTimeout {
			platform.Errorf("Health degraded after waiting %v", timer)
		}
		if platform.GetHealth().IsDegradedComparedTo(health, opts.HealthTolerance) {
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
		if k8s.IsMasterNode(node) {
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
			if currentHealth.IsDegradedComparedTo(health, opts.HealthTolerance) {
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
