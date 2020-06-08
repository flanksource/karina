package opa

import (
	"fmt"
	"path"
	"strings"

	"github.com/flanksource/karina/pkg/platform"
	minio "github.com/minio/minio-go/v6"
)

func DeployBundle(p *platform.Platform, bundlePath string) error {
	if p.OPA == nil {
		return fmt.Errorf("DeployBundle called with no OPA config specified.")
	}
	if !strings.HasSuffix(bundlePath, "tar.gz") {
		return fmt.Errorf("bundles must be a tar.gz")
	}
	s3Client, err := p.GetS3Client()
	if err != nil {
		return err
	}
	if s3Client == nil {
		return fmt.Errorf("DeployBundle failed - no S3 client found. ")
	}

	exists, err := s3Client.BucketExists(p.OPA.BundlePrefix)
	if err != nil {
		return err
	}
	if !exists {
		if err := s3Client.MakeBucket(p.OPA.BundlePrefix, p.S3.Region); err != nil {
			return err
		}
	}

	contentType := "application/x-tar"
	tarSize, err := s3Client.FPutObject(p.OPA.BundlePrefix, path.Base(bundlePath), bundlePath, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		return err
	}
	p.Debugf("Successfully uploaded %s of size %d\n", bundlePath, tarSize)

	policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::bundles/*"],"Sid": ""}]}`
	return s3Client.SetBucketPolicy(p.OPA.BundlePrefix, policy)
}
