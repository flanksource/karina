package monitoring

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	prometheusv1 "github.com/flanksource/karina/pkg/api/prometheus/v1"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	Namespace        = "monitoring"
	Prometheus       = "prometheus-k8s"
	Thanos           = "thanos"
	CaCertName       = "thanos-ca-cert"
	alertRulesSuffix = "-rules.yaml.raw"
)

const (
	ThanosObservabilityMode = "observability"
	ThanosClientMode        = "client"
)

var specs = []string{
	"prometheus-operator.yaml",
	"karma.yaml",
	"grafana-operator.yaml",
	"kube-prometheus.yaml",
	"prometheus-adapter.yaml",
	"kube-state-metrics.yaml",
	"pushgateway.yaml",
	"unmanaged/alertmanager-rules.yaml.raw",
	"unmanaged/service-monitors.yaml",
	"node-exporter.yaml",
	"alertmanager-rules.yaml.raw",
	"alertmanager-configs.yaml",
	"service-monitors.yaml",
	"namespace-rules.yaml.raw",
	"thanos/compactor.yaml",
	"thanos/querier.yaml",
	"thanos/store.yaml",
	"thanos/base.yaml",
	"kubernetes-rules.yaml.raw",
}

var monitoringNamespaceLabels = map[string]string{
	"karina.flanksource.com/namespace-name": "monitoring",
}

func Install(p *platform.Platform) error {
	if p.Monitoring.IsDisabled() {
		// setup default values so that all resources are rendered
		// so that we know what to try and delete
		for _, spec := range specs {
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

	if p.Monitoring.PushGateway.Version == "" {
		p.Monitoring.PushGateway.Version = "v1.4.1"
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

	cs := conditionalSpecs(p)
	for _, spec := range specs {
		if strings.HasSuffix(spec, alertRulesSuffix) {
			if err := deployAlertRules(p, "monitoring/"+spec); err != nil {
				return err
			}
		} else {
			fn, found := cs[spec]
			if found && fn() {
				if err := p.DeleteSpecs(v1.NamespaceAll, "monitoring/"+spec); err != nil {
					p.Errorf("failed to delete conditional spec %s: %v", spec, err)
				}
				continue
			}
			if err := p.ApplySpecs("", "monitoring/"+spec); err != nil {
				return err
			}
		}
	}

	if !p.Thanos.IsDisabled() {
		if p.Thanos.Mode != ThanosClientMode && p.Thanos.Mode != ThanosObservabilityMode {
			return fmt.Errorf("invalid thanos mode '%s',  valid options are  'client' or 'observability'", p.Thanos.Mode)
		}
		if !p.Thanos.SkipCreateBucket {
			if err := p.GetOrCreateBucket(p.Thanos.Bucket); err != nil {
				return err
			}
		}
	}

	if err := deployDashboards(p); err != nil {
		return err
	}

	return nil
}

func deployDashboards(p *platform.Platform) error {
	if p.Monitoring.Grafana.SkipDashboards {
		p.Debugf("Skipping grafana dashboard deployment")
		return nil
	}
	cd := conditionalDashboards(p)
	rootPath := "monitoring/dashboards"
	dashboards, err := p.GetResourcesByDir(rootPath, "manifests")
	if err != nil {
		return fmt.Errorf("unable to find dashboards: %v", err)
	}
	for name := range dashboards {
		fn, found := cd[name]
		if found && fn() {
			if err := p.DeleteByKind("GrafanaDashboard", Namespace, name); err != nil {
				p.Errorf("failed to delete dashboard %s: %v", name, err)
			}
			continue
		}

		contents, err := p.Template(rootPath+"/"+name, "manifests")
		if err != nil {
			return errors.Wrapf(err, "failed to template: %v", name)
		}

		if err := DeployDashboard(p, "Built-In", kommons.GetDNS1192Name(name), contents); err != nil {
			return err
		}
	}

	for _, dashboard := range p.Monitoring.Grafana.CustomDashboards {
		contents, err := ioutil.ReadFile(dashboard)

		if err != nil {
			return errors.Wrapf(err, "failed to read %s", dashboard)
		}

		if err := DeployDashboard(p, "Custom", kommons.GetDNS1192Name(path.Base(dashboard)), string(contents)); err != nil {
			return errors.Wrapf(err, "failed to deploy %s", dashboard)
		}
	}
	return nil
}

func DeployDashboard(p *platform.Platform, folder, name, contents string) error {
	return p.ApplyCRD("monitoring", kommons.CRD{
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
			"customFolderName": folder,
			"name":             name,
			"json":             contents,
		},
	})
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

		if len(p.Monitoring.ExcludeAlerts) > 0 && item.GetKind() == "PrometheusRule" {
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
		rules := []prometheusv1.Rule{}
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

func conditionalDashboards(p *platform.Platform) map[string]func() bool {
	var cd = map[string]func() bool{
		"canary-checker.json.raw":                             p.CanaryChecker.IsDisabled,
		"grafana-dashboard-log-counts.json.raw":               p.LogsExporter.IsDisabled,
		"harbor-exporter.json.raw":                            p.Harbor.IsDisabled,
		"patroni.json.raw":                                    p.PostgresOperator.IsDisabled,
		"unmanaged/etcd.json":                                 func() bool { return !p.Kubernetes.IsManaged() },
		"unmanaged/grafana-dashboard-apiserver.json":          func() bool { return !p.Kubernetes.IsManaged() },
		"unmanaged/grafana-dashboard-controller-manager.json": func() bool { return !p.Kubernetes.IsManaged() },
		"unmanaged/grafana-dashboard-scheduler.json":          func() bool { return !p.Kubernetes.IsManaged() },
	}
	return cd
}

func conditionalSpecs(p *platform.Platform) map[string]func() bool {
	var cd = map[string]func() bool{
		"thanos/compactor.yaml":                 func() bool { return p.Thanos.IsDisabled() || !p.Thanos.EnableCompactor },
		"thanos/querier.yaml":                   func() bool { return p.Thanos.IsDisabled() || p.Thanos.Mode != ThanosObservabilityMode },
		"thanos/store.yaml":                     func() bool { return p.Thanos.IsDisabled() || p.Thanos.Mode != ThanosObservabilityMode },
		"thanos/base.yaml":                      p.Thanos.IsDisabled,
		"pushgateway.yaml":                      p.Monitoring.PushGateway.IsDisabled,
		"unmanaged/alertmanager-rules.yaml.raw": func() bool { return !p.Kubernetes.IsManaged() },
		"unmanaged/service-monitors.yaml":       func() bool { return !p.Kubernetes.IsManaged() },
	}
	return cd
}
