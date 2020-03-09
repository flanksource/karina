package opa

import (
	"fmt"

	"github.com/minio/minio-go/v6"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func DeployBundle(p *platform.Platform, bundleName string) error {
	objectName := fmt.Sprintf("%s.tar.gz", bundleName)
	objectPath := fmt.Sprintf("test/opa/bundles/%s", objectName)
	log.Printf("Starting Bundle Upload")
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
	tar_size, err := s3Client.FPutObject(p.OPA.BundlePrefix, objectName, objectPath, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, tar_size)

	policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::bundles/*"],"Sid": ""}]}`
	err = s3Client.SetBucketPolicy(p.OPA.BundlePrefix, policy)
	if err != nil {
		return err
	}
	log.Printf("Bucket Policy Updated")
	return err
}
