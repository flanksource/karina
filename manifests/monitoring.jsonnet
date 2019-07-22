local mixin = import 'kube-prometheus/kube-prometheus-config-mixins.libsonnet';
local kp =
  (import 'kube-prometheus/kube-prometheus.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-kubeadm.libsonnet') +
  (import 'kube-prometheus/kube-prometheus-anti-affinity.libsonnet') +
  // (import 'kube-prometheus/kube-prometheus-managed-cluster.libsonnet') +
  // (import 'kube-prometheus/kube-prometheus-node-ports.libsonnet') +
  // (import 'kube-prometheus/kube-prometheus-static-etcd.libsonnet') +
  // (import 'kube-prometheus/kube-prometheus-thanos.libsonnet') +
  {
    _config+:: {
      namespace: 'monitoring',
       versions+:: {
        grafana: '6.2.2',
        alertmanager: "v0.17.0",
        nodeExporter: "v0.18.1",
        kubeStateMetrics: "v1.5.0",
        kubeRbacProxy: "v0.4.1",
        addonResizer: "1.8.4",
        prometheusOperator: "v0.30.0",
        prometheus: "v2.10.0",
       },

    prometheus+:: {
        names: 'k8s',
        replicas: 2,
        rules: {},
    },

    alertmanager+:: {
      name: 'main',
      config: |||
        global:
          resolve_timeout: 5m
        route:
          group_by: ['job']
          group_wait: 30s
          group_interval: 5m
          repeat_interval: 12h
          receiver: 'null'
          routes:
          - match:
              alertname: Watchdog
            receiver: 'null'
        receivers:
        - name: 'null'
      |||,
      replicas: 3,
    },

    kubeStateMetrics+:: {
      collectors: '',  // empty string gets a default set
      scrapeInterval: '30s',
      scrapeTimeout: '30s',

      baseCPU: '100m',
      baseMemory: '150Mi',
      cpuPerNode: '2m',
      memoryPerNode: '30Mi',
    },

    nodeExporter+:: {
      port: 9100,
    },
    grafana+:: {
      config: { // http://docs.grafana.org/installation/configuration/
        sections: {
          "auth.anonymous": {enabled: true},
        },
      },
    },
  }
};

{ ['00namespace-' + name]: kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus) }
{ ['0prometheus-operator-' + name]: kp.prometheusOperator[name] for name in std.objectFields(kp.prometheusOperator) } +
{ ['node-exporter-' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
{ ['kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
{ ['alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
{ ['prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) } +
{ ['grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) }
