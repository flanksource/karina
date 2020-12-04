package base

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/phases/ingress"
	"github.com/flanksource/karina/pkg/phases/platformoperator"
	"github.com/flanksource/karina/pkg/phases/quack"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"
)

func Install(platform *platform.Platform) error {
	os.Mkdir(".bin", 0755) // nolint: errcheck

	if err := platform.ApplySpecs("", "rbac.yaml"); err != nil {
		platform.Errorf("Error deploying base rbac: %s", err)
	}

	if err := platform.CreateOrUpdateNamespace("kube-system", nil, nil); err != nil {
		platform.Errorf("Error deploying base kube-system labels/annotations: %s", err)
	}

	if err := platform.CreateOrUpdateNamespace("monitoring", nil, nil); err != nil {
		platform.Errorf("Error deploying base monitoring labels/annotations: %s", err)
	}

	if err := vsphere.Install(platform); err != nil {
		return fmt.Errorf("vspere: %s", err)
	}

	if err := certmanager.Install(platform); err != nil {
		return fmt.Errorf("cert-manager: %s", err)
	}

	if err := quack.Install(platform); err != nil {
		platform.Fatalf("Error installing quack %s", err)
	}

	//platform operator has a certificate resource, wait for cert manager so that it can be fulfilled
	time.Sleep(30 * time.Second)
	platform.WaitForNamespace("cert-manager", 60*time.Second)

	if err := platformoperator.Install(platform); err != nil {
		return fmt.Errorf("platformoperator: %s", err)
	}

	if !platform.NodeLocalDNS.Disabled {
		client, err := platform.GetClientset()
		if err != nil {
			return fmt.Errorf("install: Failed to get clientset: %v", err)
		}

		kubeDNS, err := client.CoreV1().Services("kube-system").Get(context.TODO(), "kube-dns", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("install: Failed to get service: %v", err)
		}

		platform.NodeLocalDNS.DNSServer = kubeDNS.Spec.ClusterIP

		// TODO(mazzy89): make these values configurable
		platform.NodeLocalDNS.LocalDNS = "169.254.20.10"
		platform.NodeLocalDNS.DNSDomain = "cluster.local"

		if err := platform.ApplySpecs("", "node-local-dns.yaml"); err != nil {
			platform.Errorf("Error deploying node-local-dns: %s", err)
		}
	}
	if err := ingress.Install(platform); err != nil {
		return fmt.Errorf("ingress: %s", err)
	}

	if platform.LocalPath == nil || !platform.LocalPath.Disabled {
		if err := platform.CreateOrUpdateNamespace("local-path-storage", nil, nil); err != nil {
			platform.Errorf("Error creating namespace local-path-storage: %s", err)
		}
		if err := platform.ApplySpecs("", "local-path.yaml"); err != nil {
			platform.Errorf("Error deploying local path volumes: %s", err)
		}
	}

	if platform.Dashboard.Version != "" && !platform.Dashboard.Disabled {
		if err := platform.ApplySpecs("kube-system", "k8s-dashboard.yaml"); err != nil {
			platform.Errorf("Error installing K8s dashboard: %s", err)
		}
	} else {
		// set the version so that the spec is valid for deletion
		platform.Dashboard.Version = "na"
		if err := platform.DeleteSpecs("kube-system", "k8s-dashboard.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}

	if platform.NamespaceConfigurator == nil || !platform.NamespaceConfigurator.Disabled {
		if err := platform.ApplySpecs("", "namespace-configurator.yaml"); err != nil {
			platform.Errorf("Error deploying namespace configurator: %s", err)
		}
	} else {
		if err := platform.DeleteSpecs("", "namespace-configurator.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}

	if platform.S3.CSIVolumes {
		platform.Infof("Deploying S3 Volume Provisioner")
		err := platform.CreateOrUpdateSecret("csi-s3-secret", "kube-system", map[string][]byte{
			"accessKeyID":     []byte(platform.S3.AccessKey),
			"secretAccessKey": []byte(platform.S3.SecretKey),
			"endpoint":        []byte("https://" + platform.S3.Endpoint),
			"region":          []byte(platform.S3.Region),
		})
		if err != nil {
			return fmt.Errorf("install: Failed to create secret csi-s3-secret: %v", err)
		}
		if err := platform.ApplySpecs("", "csi-s3.yaml"); err != nil {
			return fmt.Errorf("install: Failed to apply specs: %v", err)
		}
	} else {
		if err := platform.DeleteSpecs("", "csi-s3.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}

	if platform.NFS != nil {
		if err := platform.ApplySpecs("", "nfs.yaml"); err != nil {
			platform.Errorf("Failed to deploy NFS %+v", err)
		}
	} else {
		if err := platform.DeleteSpecs("", "nfs.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
	}

	return nil
}
