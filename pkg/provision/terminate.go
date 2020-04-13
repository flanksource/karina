package provision

import (
	"fmt"
	"sync"
	"time"

	"github.com/moshloop/platform-cli/pkg/platform"
)

// Cleanup stops and deletes all VM's for a cluster;
func Cleanup(platform *platform.Platform) error {
	if platform.TerminationProtection {
		return fmt.Errorf("termination Protection Enabled, use -e terminationProtection=false to disable")
	}

	if err := WithVmwareCluster(platform); err != nil {
		return err
	}
	platform.Terminating = true

	vms, err := platform.Cluster.GetMachines()
	if err != nil {
		return fmt.Errorf("cleanup: failed to get VMs %v", err)
	}

	if len(vms) > platform.GetVMCount()*2 {
		platform.Fatalf("Too many VM's found, expecting +- %d but found %d", platform.GetVMCount(), len(vms))
	}

	platform.Infof("Deleting %d vm's, CTRL+C to skip, sleeping for 10s", len(vms))
	//pausing to give time for user to terminate
	time.Sleep(10 * time.Second)

	var wg sync.WaitGroup
	for _, _vm := range vms {
		vm := _vm
		wg.Add(1)
		go func() {
			defer wg.Done()
			terminate(platform, vm)
		}()
	}

	wg.Wait()
	return nil
}
