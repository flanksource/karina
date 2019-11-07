package base

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	os.Mkdir(".bin", 0755)
	if err := platform.ApplySpecs("", "base/"); err != nil {
		log.Errorf("Error deploying base stack: %s\n", err)
	}

	if platform.S3.CSIVolumes {
		log.Infof("Deploying S3 Volume Provisioner")
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

	if platform.NFS != nil {
		log.Infof("Deploying NFS Volume Provisioner: %s", platform.NFS.Host)
		if err := platform.ApplySpecs("", "nfs.yaml"); err != nil {
			log.Errorf("Failed to deploy NFS %+v", err)
		}
	}

	return nil
}
