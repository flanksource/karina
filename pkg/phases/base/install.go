package base

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	if err := platform.ApplySpecs("", "base/"); err != nil {
		log.Errorf("Error deploying base stack: %s\n", err)
	}

	if platform.S3.CSIVolumes {
		platform.CreateOrUpdateSecret("csi-s3-secret", "kube-system", map[string][]byte{
			"accessKeyID":     []byte(platform.S3.AccessKey),
			"secretAccessKey": []byte(platform.S3.SecretKey),
			"endpoint":        []byte("https://" + platform.S3.Endpoint),
			"region":          []byte(platform.S3.Region),
		})
		if err := platform.ApplySpecs("", "csi-s3.yaml"); err != nil {
			return err
		}
	}
	return nil
}
