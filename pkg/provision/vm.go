package provision

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	konfigadm "github.com/moshloop/konfigadm/pkg/types"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
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
	log.Infof("Using konfigadm spec: %s\n", konfigs)
	_vm, err := platform.Clone(*vm, konfig)

	if err != nil {
		return fmt.Errorf("vm: failed to clone %v", err)
	}
	log.Infof("Provisioned  %s ->  %s\n", vm.Name, _vm.IP)
	return nil
}
