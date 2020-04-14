package calico

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if platform.Calico.Disabled {
		platform.Debugf("Not installing calico, it is disabled")
		return nil
	}
	return platform.ApplySpecs("kube-system", "calico.yaml")
}
