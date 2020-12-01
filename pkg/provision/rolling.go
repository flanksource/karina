package provision

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/commons/timer"
	"github.com/flanksource/karina/pkg/controller/burnin"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RollingOptions struct {
	Timeout                time.Duration
	MinAge                 time.Duration
	Max                    int
	MaxSurge               int
	HealthTolerance        int
	BurninPeriod           time.Duration
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
	if kommons.IsMasterNode(node) {
		replacement, err = createSecondaryMaster(platform, opts.BurninPeriod)
		if err != nil {
			return fmt.Errorf("failed to create new secondary master: %v", err)
		}
	} else {
		replacement, err = createWorker(platform, "")
		if err != nil {
			return fmt.Errorf("failed to create new worker: %v", err)
		}
	}

	if err := waitForNode(platform, replacement.Name(), opts.Timeout); err != nil {
		platform.Errorf("[%s] terminating node that did not come up healthy", replacement)
		go func() {
			if err := cluster.Terminate(replacement); err != nil {
				platform.Errorf("failed to terminate %s, %v", replacement, err)
			}
		}()
		return err
	}
	return nil
}

func selectMachinesToReplace(platform *platform.Platform, opts RollingOptions, cluster *Cluster) *NodeMachines {
	toReplace := NodeMachines{}
	// first we select all the nodes for replacement upfront
	for _, nodeMachine := range cluster.Nodes {
		machine := nodeMachine.Machine
		node := nodeMachine.Node
		age := machine.GetAge()
		template := machine.GetTemplate()
		var newTemplate string
		if kommons.IsMasterNode(node) {
			newTemplate = platform.PlatformConfig.Master.Template
		} else {
			pool := node.Labels["karina.flanksource.com/pool"]
			newTemplate = platform.PlatformConfig.Nodes[pool].Template
		}

		// check if a node is available for update
		if kommons.IsMasterNode(node) && !opts.Masters {
			continue
		} else if !kommons.IsMasterNode(node) && !opts.Workers {
			continue
		}
		if age > opts.MinAge && newTemplate != template {
			platform.Infof("Queuing for replacement %s, age=%s, template=%s ", machine.Name(), age, template)
		} else {
			continue
		}
		toReplace.Push(nodeMachine)
	}
	sort.Sort(toReplace)
	return &toReplace
}

// Perform a rolling update of nodes
func RollingUpdate(platform *platform.Platform, opts RollingOptions) error {
	// first we start a burnin controller in the background that checks
	// new nodes with the burnin taint for health, removing the taint
	// once they become healthy
	burninCancel := make(chan bool)
	go burnin.Run(platform, opts.BurninPeriod, burninCancel)
	defer func() {
		burninCancel <- false
	}()

	// collect all info about cluster and vm's
	cluster, err := GetCluster(platform)
	if err != nil {
		return err
	}

	total := 0

	workerOpts := RollingOptions{}
	failed := false

	_ = copier.Copy(&workerOpts, &opts)
	if opts.Masters {
		opts.Workers = false
		// first we roll all masters sequentially
		opts.MaxSurge = 1
		rolled, err := roll(platform, cluster, opts)
		total += rolled
		if err != nil {
			platform.Errorf(err.Error())
			failed = true
		}
	}
	if workerOpts.Workers {
		// then we roll all the workers in batches
		workerOpts.Masters = false
		rolled, err := roll(platform, cluster, workerOpts)
		total += rolled
		if err != nil {
			platform.Errorf(err.Error())
			failed = true
		}
	}

	platform.Infof("Rollout finished, rolled %d of %d ", total, cluster.Nodes.Len())
	if failed {
		return fmt.Errorf("rolling update unsuccessful")
	}
	return nil
}

func roll(platform *platform.Platform, cluster *Cluster, opts RollingOptions) (int, error) {
	rolled := 0
	toReplace := selectMachinesToReplace(platform, opts, cluster)
	numToReplace := len(*toReplace)
	var replaced = make(chan NodeMachine, opts.MaxSurge)
	var replacementError = make(chan NodeMachine, opts.MaxSurge)
	batch := toReplace.PopN(opts.MaxSurge)
	for len(*batch) > 0 {
		platform.Infof("Replacing %s, %d remaining", *batch, toReplace.Len())
		time.Sleep(10 * time.Second)
		timer := timer.NewTimer()
		health := platform.GetHealth()
		platform.Infof("Health Before: %s", health)
		wg := sync.WaitGroup{}
		for _, nodeMachine := range *batch {
			_nodeMachine := nodeMachine
			wg.Add(1)
			go func() {
				if err := replace(platform, opts, cluster, _nodeMachine); err != nil {
					wg.Done()
					platform.Errorf(err.Error())
					replacementError <- _nodeMachine
				} else {
					wg.Done()
					replaced <- _nodeMachine
				}
			}()
			// force each replacement to have a different timestamp
			time.Sleep(2 * time.Second)
		}
		platform.Debugf("Waiting for batch to complete burn-in process")
		wg.Wait()

	outer:
		for {
			select {
			case nodeMachine := <-replacementError:
				// then we retry any errors, by putting the machine back on the queue
				toReplace.Push(nodeMachine)
			case nodeMachine := <-replaced:
				// terminate successful replacements
				// nolint: errcheck
				go cluster.Terminate(nodeMachine.Machine)
				rolled++
			default:
				platform.Debugf("Batch completed burn-in process ")
				// batch is done
				break outer
			}
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
			return rolled, fmt.Errorf("health degraded after waiting %v", timer)
		}
		if platform.GetHealth().IsDegradedComparedTo(health, opts.HealthTolerance) {
			return rolled, fmt.Errorf("cluster is not healthy, aborting rollout after %d of %d ", rolled, cluster.Nodes.Len())
		}
		if rolled >= opts.Max {
			break
		}
		// select the next batch of nodes to update
		batch = toReplace.PopN(opts.MaxSurge)
	}

	err := error(nil)
	if rolled < numToReplace {
		err = fmt.Errorf("rolling update failed to replace all scheduled nodes")
	}
	return rolled, err
}

// Perform a rolling restart of nodes
func RollingRestart(platform *platform.Platform, opts RollingOptions) error {
	client, err := platform.GetClientset()
	if err != nil {
		return err
	}

	list, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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
		if kommons.IsMasterNode(node) {
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
