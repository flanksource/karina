package nfs

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = constants.KubeSystem

func Install(platform *platform.Platform) error {
	if platform.NFS == nil {
		return platform.DeleteSpecs(Namespace, "nfs.yaml")
	}
	return platform.ApplySpecs(Namespace, "nfs.yaml")
}
