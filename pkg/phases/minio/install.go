package minio

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "minio"

func Install(platform *platform.Platform) error {
	if platform.Minio.Replicas == 0 {
		platform.Minio.Replicas = 2
	}
	if platform.Minio.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "minio.yaml")
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	return platform.ApplySpecs(Namespace, "minio.yaml")

}
