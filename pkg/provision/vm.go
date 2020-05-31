package provision

import (
	"fmt"

	konfigadm "github.com/flanksource/konfigadm/pkg/types"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

// VM provisions a new standalone VM
func VM(platform *platform.Platform, vm *types.VM, konfigs ...string) error {
	if err := WithVmwareCluster(platform); err != nil {
		return err
	}

	konfig, err := konfigadm.NewConfig(konfigs...).Build()
	if err != nil {
		return fmt.Errorf("vm: failed to get new config: %v", err)
	}
	platform.Infof("Using konfigadm spec: %s\n", konfigs)
	machine, err := platform.Clone(*vm, konfig)

	if err != nil {
		return fmt.Errorf("vm: failed to clone %v", err)
	}
	platform.Infof("Provisioned  %s ->  %s\n", machine.Name(), machine.IP())
	return nil
}
