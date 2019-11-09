package base

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	os.Mkdir(".bin", 0755)

	if err := platform.ApplySpecs("", "rbac.yml"); err != nil {
		log.Errorf("Error deploying base rbac: %s\n", err)
	}

	if platform.CertManager == nil || !platform.CertManager.Disabled {
		log.Infof("Installing CertMananager")
		if err := platform.ApplySpecs("", "cert-manager-crd.yml"); err != nil {
			log.Errorf("Error deploying cert manager CRDs: %s\n", err)
		}

		// the cert-manager webhook can take time to deploy, so we deploy it once ignoring any errors
		// wait for 180s for the namespace to be ready, deploy again (usually a no-op) and only then report errors
		var _ = platform.ApplySpecs("", "cert-manager-deploy.yml")
		platform.WaitForNamespace("cert-manager", 180*time.Second)
		if err := platform.ApplySpecs("", "cert-manager-deploy.yml"); err != nil {
			log.Errorf("Error deploying cert manager: %s\n", err)
		}
	}

	if platform.Quack == nil || !platform.Quack.Disabled {
		log.Infof("Installing Quack")
		if err := platform.ApplySpecs("", "quack.yml"); err != nil {
			log.Errorf("Error deploying quack: %s\n", err)
		}
	}

	if platform.LocalPath == nil || !platform.LocalPath.Disabled {
		log.Infof("Installing local path volumes")
		if err := platform.ApplySpecs("", "local-path.yml"); err != nil {
			log.Errorf("Error deploying local path volumes: %s\n", err)
		}
	}

	if platform.Dashboard == nil || !platform.Dashboard.Disabled {
		log.Infof("Installing K8s dashboard")
		if err := platform.ApplySpecs("", "k8s-dashboard.yml"); err != nil {
			log.Errorf("Error K8s dashboard: %s\n", err)
		}
	}

	if platform.NamespaceConfigurator == nil || !platform.NamespaceConfigurator.Disabled {
		log.Infof("Installing namespace configurator")
		if err := platform.ApplySpecs("", "namespace-configurator.yml"); err != nil {
			log.Errorf("Error deploying namespace configurator: %s\n", err)
		}
	}

	if platform.PlatformOperator == nil || platform.PlatformOperator.Disabled {
		log.Infof("Installing platform operator")
		if err := platform.ApplySpecs("", "platform-operator.yml"); err != nil {
			log.Errorf("Error deploying platform-operator: %s\n", err)
		}
	}

	if platform.Nginx == nil || !platform.Nginx.Disabled {
		log.Infof("Installing Nginx Ingress Controller")
		if err := platform.ApplySpecs("", "nginx.yml"); err != nil {
			log.Errorf("Error deploying nginx: %s\n", err)
		}
	}

	if platform.Minio == nil || !platform.Minio.Disabled {
		log.Infof("Installing minio")
		if err := platform.ApplySpecs("", "minio.yaml"); err != nil {
			log.Errorf("Error deploying minio: %s\n", err)
		}
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

	if platform.Flux != nil && !platform.Flux.Disabled {
		if platform.Flux.Version == "" {
			platform.Flux.Version = "1.15.0"

		}
		if platform.Flux.Image == "" {
			platform.Flux.Image = "docker.io/fluxcd/flux"
		}
		log.Infof("Deploying Flux %s", platform.Flux.Version)
		if err := platform.ApplySpecs("", "flux.yml"); err != nil {
			log.Errorf("Failed to deploy flux %+v", err)
		}
	}

	return nil
}
