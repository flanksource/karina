package monitoring

import (
	"encoding/json"
	"fmt"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"

	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "monitoring"
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

	urls := map[string]string {
		"alertmanager": fmt.Sprintf("https://alertmanager.%s", p.Domain),
		"grafana": fmt.Sprintf("https://grafana.%s", p.Domain),
		"prometheus": fmt.Sprintf("https://prometheus.%s", p.Domain),
	}

	for name, file := range dashboards {
		contents := text.SafeRead(file)
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

	return nil
}
