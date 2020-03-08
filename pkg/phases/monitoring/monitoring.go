package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana-tools/sdk"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

const (
	Namespace       = "monitoring"
	StoreCertName   = "thanos-store-grpc"
	SidecarCertName = "thanos-sidecar"
	QueryCertName   = "thanos-query"
	CaCertName      = "thanos-ca-cert"
)

var specs = []string{"grafana-operator.yml", "kube-prometheus.yml", "prometheus-adapter.yml", "kube-state-metrics.yml", "node-exporter.yml", "alertmanager-rules.yml.raw", "service-monitors.yml"}

func Install(p *platform.Platform) error {
	if p.Monitoring == nil || p.Monitoring.Disabled {
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := p.ApplySpecs("", "monitoring/prometheus-operator.yml"); err != nil {
		log.Warnf("Failed to deploy prometheus operator %v", err)
	}

	data, err := p.Template("monitoring/alertmanager.yaml", "manifests")
	if err != nil {
		return fmt.Errorf("install: failed to template alertmanager manifests: %v", err)
	}
	if err := p.CreateOrUpdateSecret("alertmanager-main", Namespace, map[string][]byte{
		"alertmanager.yaml": []byte(data),
	}); err != nil {
		return fmt.Errorf("install: failed to create/update secret: %v", err)
	}

	for _, spec := range specs {
		log.Infof("Applying %s", spec)
		if err := p.ApplySpecs("", "monitoring/"+spec); err != nil {
			return fmt.Errorf("install: failed to apply monitoring specs: %v", err)
		}
	}

	dashboards, err := p.GetResourcesByDir("/monitoring/dashboards", "manifests")
	if err != nil {
		return fmt.Errorf("Unable to find dashboards: %v", err)
	}

	urls := map[string]string{
		"alertmanager": fmt.Sprintf("https://alertmanager.%s", p.Domain),
		"grafana":      fmt.Sprintf("https://grafana.%s", p.Domain),
		"prometheus":   fmt.Sprintf("https://prometheus.%s", p.Domain),
	}

	for name := range dashboards {
		contents, err := p.Template("/monitoring/dashboards/"+name, "manifests")
		if err != nil {
			fmt.Errorf("Failed to template the dashboard: %v ", err)
		}
		var board sdk.Board
		if err := json.Unmarshal([]byte(contents), &board); err != nil {
			log.Warnf("Invalid grafana dashboard %s: %v", name, err)
		}

		for i := range board.Templating.List {
			for k, v := range urls {
				if k == board.Templating.List[i].Name {
					board.Templating.List[i].Current.Value = v
					board.Templating.List[i].Current.Text = v
					board.Templating.List[i].Query = v
				}
			}
		}

		contentsModified, err := json.Marshal(&board)
		if err != nil {
			log.Warnf("Failed to marshal dashboard json %s: %v", name, err)
		}

		if err := p.ApplyCRD("monitoring", k8s.CRD{
			ApiVersion: "integreatly.org/v1alpha1",
			Kind:       "GrafanaDashboard",
			Metadata: k8s.Metadata{
				Name:      name,
				Namespace: Namespace,
				Labels: map[string]string{
					"app": "grafana",
				},
			},
			Spec: map[string]interface{}{
				"name": name,
				"json": string(contentsModified),
			},
		}); err != nil {
			return fmt.Errorf("install: failed to apply CRD: %v", err)
		}
	}

	return deployThanos(p)
}

func deployThanos(p *platform.Platform) error {
	if p.Thanos == nil || p.Thanos.Disabled {
		log.Debugln("Thanos is disabled")
		return nil
	}

	s3Client, err := p.GetS3Client()
	if err != nil {
		return err
	}

	exists, err := s3Client.BucketExists(p.Thanos.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s3Client.MakeBucket(p.Thanos.Bucket, p.S3.Region); err != nil {
			return err
		}
	}
	if err := p.ApplySpecs("", "monitoring/thanos-config.yaml"); err != nil {
		return err
	}
	if err := p.ApplySpecs("", "monitoring/thanos-sidecar.yaml"); err != nil {
		return err
	}
	caCert := p.ReadIngressCACertString()
	content := map[string][]byte{
		"ca.crt": []byte(caCert),
	}
	if err := p.CreateOrUpdateSecret(CaCertName, Namespace, content); err != nil {
		return fmt.Errorf("install: failed to create secret with CA certificate: %v", err)
	}

	if p.Thanos.Mode == "client" {
		log.Info("Thanos in client mode is enabled. Sidecar will be deployed within Promerheus pod.")
		if p.Thanos.ThanosSidecarEndpoint == "" || p.Thanos.ThanosSidecarPort == "" {
			return errors.New("thanosSidecarEndpoint and thanosSidecarPort should not be empty in client mode")
		}
		if !p.HasSecret(Namespace, SidecarCertName) {
			cert, err := p.CreateIngressCertificate("thanos-sidecar")
			if err != nil {
				return fmt.Errorf("install: failed to create ingress certificate: %v", err)
			}
			if err := p.CreateOrUpdateSecret(SidecarCertName, Namespace, cert.AsTLSSecret()); err != nil {
				return fmt.Errorf("install: failed to create secret with certificate and key: %v", err)
			}
		}
		p.ApplySpecs("", "monitoring/thanos-sidecar.yaml")
	} else if p.Thanos.Mode == "observability" {
		log.Info("Thanos in observability mode is enabled. Compactor, Querier and Store will be deployed.")
		if p.Thanos.ThanosSidecarEndpoint != "" || p.Thanos.ThanosSidecarPort != "" {
			return errors.New("thanosSidecarEndpoint and thanosSidecarPort are not empty. Please use clientSidecars to specify client sidecars")
		}
		thanosCerts := []string{StoreCertName, SidecarCertName, QueryCertName}
		for _, certName := range thanosCerts {
			if !p.HasSecret(Namespace, certName) {
				cert, err := p.CreateInternalCertificate(certName, "monitoring", "cluster.local")
				if err != nil {
					return fmt.Errorf("install: failed to create internal certificate: %v", err)
				}
				if err := p.CreateOrUpdateSecret(certName, Namespace, cert.AsTLSSecret()); err != nil {
					return fmt.Errorf("install: failed to create secret with certificate and key for %s: %v", certName, err)
				}
			}
		}
		thanosSpecs := []string{"thanos-compactor.yaml", "thanos-querier.yaml", "thanos-store.yaml"}
		for _, spec := range thanosSpecs {
			log.Infof("Applying %s", spec)
			if err := p.ApplySpecs("", "monitoring/observability/"+spec); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("invalid thanos mode '%s',  valid options are  'client' or 'observability'", p.Thanos.Mode)
	}
	return nil

}
