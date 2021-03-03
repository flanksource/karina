package s3

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = constants.KubeSystem

func Install(platform *platform.Platform) error {
	if !platform.S3.CSIVolumes {
		return platform.DeleteSpecs(Namespace, "csi-s3.yaml")
	}

	err := platform.CreateOrUpdateSecret("csi-s3-secret", Namespace, map[string][]byte{
		"accessKeyID":     []byte(platform.S3.AccessKey),
		"secretAccessKey": []byte(platform.S3.SecretKey),
		"endpoint":        []byte("https://" + platform.S3.Endpoint),
		"region":          []byte(platform.S3.Region),
	})
	if err != nil {
		return err
	}
	return platform.ApplySpecs(Namespace, "csi-s3.yaml")
}
