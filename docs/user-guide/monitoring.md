The monitoring stack includes the following components, which are all accessible via default ingress addresses



| Component                                                    | Use                                                    | Access                              |
| ------------------------------------------------------------ | ------------------------------------------------------ | ----------------------------------- |
| Prometheus                                                   |                                                        | https://prometheus.CLUSTER_DOMAIN   |
| Alert Manager                                                |                                                        | https://alertmanager.CLUSTER_DOMAIN |
| [prometheus-operator](https://coreos.com/operators/prometheus/docs/latest/api.html) | Declarative management of prometheus configuration     |                                     |
| Grafana                                                      | Monitoring Dashboards                                  | https://grafana.CLUSTER_DOMAIN      |
| [Grafana Operator](https://github.com/integr8ly/grafana-operator/) | Declarative/GitOps configuration of grafana dashboards |                                     |
| [Thanos](https://thanos.io/)                                 |                                                        | https://thanos.CLUSTER_DOMAIN       |
| Canary Checker                                               |                                                        | https://canaries.CLUSTER_DOMAIN     |

### Scraping metrics

In order to ingest metrics you need to configure a service monitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name:  YOUR-APP-MONITOR
  namespace: YOUR-NAMESPACE
spec:
  endpoints:
    - interval: 30s
      targetPort: 8080
  jobLabel: canary-checker
  namespaceSelector:
    matchNames:
      - YOUR-NAMESPACE
  selector:
    matchLabels:
      app: YOUR-APP
```

See [ServiceMonitorSpec](https://coreos.com/operators/prometheus/docs/latest/api.html#servicemonitorspec) for all available options

### Creating custom dashboards

!!! warning
    The default password for Grafana is `root`/`secret` but note that any changes will not persist across restarts. To make changes persistent export the dashboard as JSON
    and create a `GrafanaDashboard` CRD

```yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: simple-dashboard
  labels:
    app: grafana
spec:
  name: simple-dashboard.json
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```