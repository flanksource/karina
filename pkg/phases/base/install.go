package base

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/phases/ingress"
	"github.com/moshloop/platform-cli/pkg/phases/nginx"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform) error {
	os.Mkdir(".bin", 0755)

	if err := platform.ApplySpecs("", "rbac.yml"); err != nil {
		log.Errorf("Error deploying base rbac: %s\n", err)
	}

	if err := platform.ApplySpecs("", "tiller.yml"); err != nil {
		log.Errorf("Error deploying tiller: %s\n", err)
	}

	if !platform.NodeLocalDNS.Disabled {
		client, err := platform.GetClientset()
		if err != nil {
			return fmt.Errorf("install: Failed to get clientset: %v", err)
		}

		kubeDNS, err := client.CoreV1().Services("kube-system").Get("kube-dns", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("install: Failed to get service: %v", err)
		}

		platform.NodeLocalDNS.DNSServer = kubeDNS.Spec.ClusterIP

		// TODO(mazzy89): make these values configurable
		platform.NodeLocalDNS.LocalDNS = "169.254.20.10"
		platform.NodeLocalDNS.DNSDomain = "cluster.local"

		if err := platform.ApplySpecs("", "node-local-dns.yml"); err != nil {
			log.Errorf("Error deploying node-local-dns: %s\n", err)
		}
	}

	if platform.CertManager == nil || !platform.CertManager.Disabled {
		log.Infof("Installing CertMananager")
		if err := platform.ApplySpecs("", "cert-manager-crd.yml"); err != nil {
			log.Errorf("Error deploying cert manager CRDs: %s\n", err)
		}

		// the cert-manager webhook can take time to deploy, so we deploy it once ignoring any errors
		// wait for 180s for the namespace to be ready, deploy again (usually a no-op) and only then report errors
		var _ = platform.ApplySpecs("", "cert-manager-deploy.yml")
		platform.GetKubectl()("wait --for=condition=Available apiservice v1beta1.webhook.cert-manager.io")
		platform.WaitForNamespace("cert-manager", 180*time.Second)
		if err := platform.ApplySpecs("", "cert-manager-deploy.yml"); err != nil {
			log.Errorf("Error deploying cert manager: %s\n", err)
		}
	}

	if err := platform.CreateOrUpdateNamespace(constants.PlatformSystem, map[string]string{
		"quack.pusher.com/enabled": "true",
	}, nil); err != nil {
		return err
	}

	var secrets = make(map[string][]byte)

	secrets["AWS_ACCESS_KEY_ID"] = []byte(platform.S3.AccessKey)
	secrets["AWS_SECRET_ACCESS_KEY"] = []byte(platform.S3.SecretKey)

	if platform.Ldap != nil {
		secrets["LDAP_USERNAME"] = []byte(platform.Ldap.Username)
		secrets["LDAP_PASSWORD"] = []byte(platform.Ldap.Password)
	}

	if err := platform.CreateOrUpdateSecret("secrets", constants.PlatformSystem, secrets); err != nil {
		return err
	}

	if err := nginx.Install(platform); err != nil {
		log.Fatalf("Error deploying nginx %s\n", err)
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

	if !platform.Dashboard.Disabled {
		log.Infof("Installing K8s dashboard")
		platform.Dashboard.AccessRestricted.Snippet = ingress.IngressNginxAccessSnippet(platform, platform.Dashboard.AccessRestricted)
		if err := platform.ApplySpecs("", "k8s-dashboard.yml"); err != nil {
			log.Errorf("Error installing K8s dashboard: %s\n", err)
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
			return fmt.Errorf("install: Failed to apply specs: %v", err)
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
