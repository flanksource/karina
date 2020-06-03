package s3uploadcleaner

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.S3UploadCleaner.Disabled {
		p.Infof("Skipping deployment of s3-upload-cleaner, it is disabled")
		return nil
	}

	return p.ApplySpecs(constants.PlatformSystem, "s3-upload-cleaner.yaml")
}
