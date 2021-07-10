package monitoring

import (
	"encoding/json"
	"fmt"
	"strings"

	prometheusv1 "github.com/flanksource/karina/pkg/api/prometheus/v1"
	v1 "github.com/flanksource/karina/pkg/api/prometheus/v1"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	Namespace        = "monitoring"
	Prometheus       = "prometheus-k8s"
	Thanos           = "thanos"
	CaCertName       = "thanos-ca-cert"
	alertRulesSuffix = "-rules.yaml.raw"
)

var specs = []string{
	"karma.yaml",
	"grafana-operator.yaml",
	"kube-prometheus.yaml",
	"prometheus-adapter.yaml",
	"kube-state-metrics.yaml",
	"node-exporter.yaml",
	"alertmanager-rules.yaml.raw",
	"alertmanager-configs.yaml",
	"service-monitors.yaml",
	"namespace-rules.yaml.raw",
	"kubernetes-rules.yaml.raw",
}

var unmanagedSpecs = []string{
	"alertmanager-rules.yaml.raw",
	"service-monitors.yaml",
}

var cleanup = []string{
	"observability/thanos-compactor.yaml",
	"observability/thanos-querier.yaml",
	"observability/thanos-store.yaml",
	"thanos-config.yaml",
}

var monitoringNamespaceLabels = map[string]string{
	"karina.flanksource.com/namespace-name": "monitoring",
}

func Install(p *platform.Platform) error {
	if p.Monitoring.IsDisabled() {
		// setup default values so that all resources are rendered
		// so that we know what to try and delete
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

	if err := p.CreateOrUpdateNamespace(Namespace, monitoringNamespaceLabels, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if !p.HasSecret(Namespace, "alertmanager-relabeling") {
		if err := p.CreateOrUpdateSecret("alertmanager-relabeling", Namespace, map[string][]byte{
			"config.yaml": nil,
		}); err != nil {
			return nil
		}
	}
	if err := p.CreateOrUpdateSecret(CaCertName, Namespace, map[string][]byte{
		"ca.crt": p.GetIngressCA().GetPublicChain()[0].EncodedCertificate(),
	}); err != nil {
		return fmt.Errorf("install: failed to create secret with CA certificate: %v", err)
	}

	if err := p.ApplySpecs("", "monitoring/prometheus-operator.yaml"); err != nil {
		p.Warnf("Failed to deploy prometheus operator %v", err)
	}

	for _, spec := range specs {
		if strings.HasSuffix(spec, alertRulesSuffix) {
			if err := deployAlertRules(p, "monitoring/"+spec); err != nil {
				return fmt.Errorf("install: failed to deploy alert rules %s: %v", spec, err)
			}
		} else {
			if err := p.ApplySpecs("", "monitoring/"+spec); err != nil {
				return fmt.Errorf("install: failed to apply monitoring specs: %v", err)
			}
		}
	}

	if !p.Kubernetes.Managed {
		for _, spec := range unmanagedSpecs {
			if strings.HasSuffix(spec, alertRulesSuffix) {
				if err := deployAlertRules(p, "monitoring/unmanaged/"+spec); err != nil {
					return fmt.Errorf("install: failed to deploy alert rules %s: %v", spec, err)
				}
			} else {
				if err := p.ApplySpecs("", "monitoring/unmanaged/"+spec); err != nil {
					return fmt.Errorf("install: failed to apply monitoring specs: %v", err)
				}
			}
		}
	}

	err := deployDashboards(p, "monitoring/dashboards")
	if err != nil {
		return err
	}

	if !p.Kubernetes.Managed {
		err = deployDashboards(p, "monitoring/dashboards/unmanaged")
		if err != nil {
			return err
		}
	}

	return deployThanos(p)
}

func deployDashboards(p *platform.Platform, rootPath string) error {
	dashboards, err := p.GetResourcesByDir(rootPath, "manifests")
	if err != nil {
		return fmt.Errorf("unable to find dashboards: %v", err)
	}
	for name := range dashboards {
		if err := DeployDashboard(p, name, rootPath+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

func DeployDashboard(p *platform.Platform, name, file string) error {
	contents, err := p.Template(file, "manifests")
	if err != nil {
		return fmt.Errorf("failed to template the dashboard: %v ", err)
	}
	if err := p.ApplyCRD("monitoring", kommons.CRD{
		APIVersion: "integreatly.org/v1alpha1",
		Kind:       "GrafanaDashboard",
		Metadata: kommons.Metadata{
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

	return nil
}

func deployAlertRules(p *platform.Platform, spec string) error {
	template, err := p.Template(spec, "manifests")
	if err != nil {
		return errors.Wrapf(err, "failed to template manifests: %v", spec)
	}

	items, err := kommons.GetUnstructuredObjects([]byte(template))
	if err != nil {
		return err
	}

	for _, item := range items {
		obj := item

		if item.GetKind() == "PrometheusRule" {
			jsonObj := kommons.ToJson(item)
			prometheusRule := &prometheusv1.PrometheusRule{}
			if err := json.Unmarshal([]byte(jsonObj), prometheusRule); err != nil {
				return errors.Wrap(err, "failed to unmarshal prometheus rule")
			}

			updatedRule := filterPrometheusRules(p, prometheusRule)
			o, err := kommons.ToUnstructured(&unstructured.Unstructured{}, updatedRule)
			if err != nil {
				return errors.Wrap(err, "failed to convert prometheus rule to unstructured")
			}
			obj = o
		}

		if err := p.ApplyUnstructured("", obj); err != nil {
			return err
		}
	}

	return nil
}

func filterPrometheusRules(p *platform.Platform, rule *prometheusv1.PrometheusRule) *prometheusv1.PrometheusRule {
	newRule := rule.DeepCopy()

	excludedRules := map[string]bool{}
	for _, rule := range p.Monitoring.ExcludeAlerts {
		excludedRules[rule] = true
	}

	ruleGroups := []prometheusv1.RuleGroup{}

	for _, group := range newRule.Spec.Groups {
		rules := []v1.Rule{}
		for _, rule := range group.Rules {
			if rule.Alert != "" {
				_, found := excludedRules[rule.Alert]
				if found {
					p.Debugf("excluding alert rule %s", rule.Alert)
				} else {
					rules = append(rules, rule)
				}
			} else {
				rules = append(rules, rule)
			}
		}
		if len(rules) > 0 {
			group.Rules = rules
			ruleGroups = append(ruleGroups, group)
		}
	}

	newRule.Spec.Groups = ruleGroups

	return newRule
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

	if p.Thanos.Mode == "client" {
		//Thanos in client mode is enabled. Sidecar will be deployed within prometheus pod
	} else if p.Thanos.Mode == "observability" {
		// Thanos in observability mode is enabled. Compactor, Querier and Store will be deployed
		thanosSpecs := []string{"thanos-querier.yaml", "thanos-store.yaml"}
		for _, spec := range thanosSpecs {
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
