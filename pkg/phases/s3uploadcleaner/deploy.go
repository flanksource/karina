package s3uploadcleaner

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.S3UploadCleaner == nil || p.S3UploadCleaner.Disabled {
		return p.DeleteSpecs(constants.PlatformSystem, "s3-upload-cleaner.yaml")
	}

	return p.ApplySpecs(constants.PlatformSystem, "s3-upload-cleaner.yaml")
}
