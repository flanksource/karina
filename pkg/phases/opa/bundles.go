package opa

import (
	"fmt"

	"github.com/flanksource/commons/exec"
	"github.com/minio/minio-go/v6"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func DeployBundles(p *platform.Platform, bundlesPath string) error {
	objectName := fmt.Sprintf("%s.tar.gz", p.OPA.BundleServiceName)
	tar_command := fmt.Sprintf("tar -C  %s -czvf %s  .manifest %s", bundlesPath, objectName, p.OPA.BundleServiceName)
	err := exec.Exec(tar_command)
	if err == nil {
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
		tar_size, err := s3Client.FPutObject(p.OPA.BundlePrefix, objectName, objectName, minio.PutObjectOptions{ContentType: contentType})

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
	}
	return err
}
