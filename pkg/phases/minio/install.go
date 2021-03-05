package minio

import (
	"time"

	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "minio"

func Install(platform *platform.Platform) error {
	if platform.Minio.Replicas == 0 {
		platform.Minio.Replicas = 1
	}
	if platform.Minio.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "minio.yaml")
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "minio.yaml"); err != nil {
		return nil
	}

	return platform.WaitForNamespace(Namespace, 60*time.Second)
}
