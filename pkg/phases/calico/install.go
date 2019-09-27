package calico

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	return platform.ApplySpecs("kube-system", "calico.yaml")
}
