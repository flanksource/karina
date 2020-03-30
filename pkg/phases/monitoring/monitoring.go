package monitoring

import (
	"encoding/json"
	"fmt"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace  = "monitoring"
	CaCertName = "thanos-ca-cert"
)

var specs = []string{
	"grafana-operator.yaml",
	"kube-prometheus.yaml",
	"prometheus-adapter.yaml",
	"kube-state-metrics.yaml",
	"node-exporter.yaml",
	"alertmanager-rules.yaml.raw",
	"service-monitors.yaml",
}

func Install(p *platform.Platform) error {
	if p.Monitoring == nil || p.Monitoring.Disabled {
		return nil
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
		return fmt.Errorf("unable to find dashboards: %v", err)
	}

	urls := map[string]string{
		"alertmanager": fmt.Sprintf("https://alertmanager.%s", p.Domain),
		"grafana":      fmt.Sprintf("https://grafana.%s", p.Domain),
		"prometheus":   fmt.Sprintf("https://prometheus.%s", p.Domain),
	}

	for name := range dashboards {
		contents, err := p.Template("/monitoring/dashboards/"+name, "manifests")
		if err != nil {
			return fmt.Errorf("failed to template the dashboard: %v ", err)
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

	if p.S3.ExternalEndpoint == "" {
		p.S3.ExternalEndpoint = p.S3.Endpoint
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
		log.Info("Thanos in client mode is enabled. Sidecar will be deployed within prometheus pod.")
	} else if p.Thanos.Mode == "observability" {
		log.Info("Thanos in observability mode is enabled. Compactor, Querier and Store will be deployed.")
		thanosSpecs := []string{"thanos-querier.yaml", "thanos-store.yaml"}
		for _, spec := range thanosSpecs {
			log.Infof("Applying %s", spec)
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
