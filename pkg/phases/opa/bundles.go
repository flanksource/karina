package opa

import (
	"fmt"

	minio "github.com/minio/minio-go/v6"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func DeployBundle(p *platform.Platform, bundleName string) error {
	objectName := fmt.Sprintf("%s.tar.gz", bundleName)
	objectPath := fmt.Sprintf("test/opa/bundles/%s", objectName)
	s3Client, err := p.GetS3Client()
	if err != nil {
		return err
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
	tarSize, err := s3Client.FPutObject(p.OPA.BundlePrefix, objectName, objectPath, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		return err
	}
	p.Debugf("Successfully uploaded %s of size %d\n", objectName, tarSize)

	policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::bundles/*"],"Sid": ""}]}`
	return s3Client.SetBucketPolicy(p.OPA.BundlePrefix, policy)
}
