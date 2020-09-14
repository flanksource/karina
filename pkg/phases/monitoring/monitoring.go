package monitoring

import (
	"fmt"

	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

const (
	Namespace  = "monitoring"
	CaCertName = "thanos-ca-cert"
)

var specs = []string{
	"karma.yaml",
	"grafana-operator.yaml",
	"kube-prometheus.yaml",
	"prometheus-adapter.yaml",
	"kube-state-metrics.yaml",
	"node-exporter.yaml",
	"alertmanager-rules.yaml.raw",
	"service-monitors.yaml",
	"namespace-rules.yaml.raw",
	"kubernetes-rules.yaml.raw",
}

var cleanup = []string{
	"observability/thanos-compactor.yaml",
	"observability/thanos-querier.yaml",
	"observability/thanos-store.yaml",
	"thanos-config.yaml",
	"thanos-sidecar.yaml",
}

func Install(p *platform.Platform) error {
	if p.Monitoring == nil || p.Monitoring.Disabled {
		// setup default values so that all resources are rendered
		// so that we know what to try and delete
		p.Monitoring = &types.Monitoring{}
		p.Thanos = &types.Thanos{Mode: "observability"}
		p.Thanos.Version = "deleted"
		for _, spec := range append(specs, cleanup...) {
			if err := p.DeleteSpecs(Namespace, "monitoring/"+spec); err != nil {
				p.Warnf("failed to delete specs: %v", err)
			}
		}
		return nil
	}

	if p.Monitoring.Karma.Version == "" {
		p.Monitoring.Karma.Version = "v0.63"
	}

	if p.Monitoring.Karma.AlertManagers == nil {
		p.Monitoring.Karma.AlertManagers = map[string]string{}
	}

	if len(p.Monitoring.Karma.AlertManagers) == 0 {
		p.Monitoring.Karma.AlertManagers["alertmanager-main"] = "http://alertmanager-main:9093"
	}
	if p.Monitoring.Prometheus.Version == "" {
		p.Monitoring.Prometheus.Version = "v2.19.0"
	}

	if p.Monitoring.AlertManager.Version == "" {
		p.Monitoring.AlertManager.Version = "v0.20.0"
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := p.CreateOrUpdateSecret(CaCertName, Namespace, map[string][]byte{
		"ca.crt": p.GetIngressCA().GetPublicChain()[0].EncodedCertificate(),
	}); err != nil {
		return fmt.Errorf("install: failed to create secret with CA certificate: %v", err)
	}

	if err := p.ApplySpecs("", "monitoring/prometheus-operator.yaml"); err != nil {
		p.Warnf("Failed to deploy prometheus operator %v", err)
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
		if err := p.ApplySpecs("", "monitoring/"+spec); err != nil {
			return fmt.Errorf("install: failed to apply monitoring specs: %v", err)
		}
	}

	dashboards, err := p.GetResourcesByDir("/monitoring/dashboards", "manifests")
	if err != nil {
		return fmt.Errorf("unable to find dashboards: %v", err)
	}
	for name := range dashboards {
		contents, err := p.Template("/monitoring/dashboards/"+name, "manifests")
		if err != nil {
			return fmt.Errorf("failed to template the dashboard: %v ", err)
		}
		if err := p.ApplyCRD("monitoring", k8s.CRD{
			APIVersion: "integreatly.org/v1alpha1",
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
				"json": contents,
			},
		}); err != nil {
			return fmt.Errorf("install: failed to apply CRD: %v", err)
		}
	}

	return deployThanos(p)
}

func deployThanos(p *platform.Platform) error {
	if p.Thanos == nil || p.Thanos.IsDisabled() {
		return nil
	}

	if err := p.GetOrCreateBucket(p.Thanos.Bucket); err != nil {
		return err
	}

	if err := p.ApplySpecs("", "monitoring/thanos-config.yaml"); err != nil {
		return err
	}
	if err := p.ApplySpecs("", "monitoring/thanos-sidecar.yaml"); err != nil {
		return err
	}

	if p.Thanos.Mode == "client" {
		p.Infof("Thanos in client mode is enabled. Sidecar will be deployed within prometheus pod.")
	} else if p.Thanos.Mode == "observability" {
		p.Infof("Thanos in observability mode is enabled. Compactor, Querier and Store will be deployed.")
		thanosSpecs := []string{"thanos-querier.yaml", "thanos-store.yaml"}
		for _, spec := range thanosSpecs {
			p.Infof("Applying %s", spec)
			if err := p.ApplySpecs("", "monitoring/observability/"+spec); err != nil {
				return err
			}
		}
		if p.Thanos.EnableCompactor {
			if err := p.ApplySpecs("", "monitoring/observability/thanos-compactor.yaml"); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("invalid thanos mode '%s',  valid options are  'client' or 'observability'", p.Thanos.Mode)
	}

	return nil
}
