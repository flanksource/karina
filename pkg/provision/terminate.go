package provision

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

// Cleanup stops and deletes all VM's for a cluster;
func Cleanup(platform *platform.Platform) error {
	if platform.TerminationProtection {
		return fmt.Errorf("termination Protection Enabled, use -e terminationProtection=false to disable")
	}

	if err := platform.OpenViaEnv(); err != nil {
		return fmt.Errorf("cleanup: failed to open via env %v", err)
	}

	vms, err := platform.GetVMs()
	if err != nil {
		return fmt.Errorf("cleanup: failed to get VMs %v", err)
	}

	if len(vms) > platform.GetVMCount()*2 {
		log.Fatalf("Too many VM's found, expecting +- %d but found %d", platform.GetVMCount(), len(vms))
	}

	log.Infof("Deleting %d vm's, CTRL+C to skip, sleeping for 10s", len(vms))
	//pausing to give time for user to terminate
	time.Sleep(10 * time.Second)

	var wg sync.WaitGroup
	for _, _vm := range vms {
		vm := _vm
		if platform.DryRun {
			continue
		}
		wg.Add(1)
		go func() {
			vm.Terminate() // nolint: errcheck
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}
