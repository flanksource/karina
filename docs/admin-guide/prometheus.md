

The monitoring stack includes the following components, which are all accessible via default ingress addresses

| Component                                                    | Use                                                    | Access                         |
| ------------------------------------------------------------ | ------------------------------------------------------ | ------------------------------ |
| :prometheus: Prometheus                                                   |                                                        | [:octicons-link-external-24: prometheus.%%{domain}%%](https://prometheus.%%{domain}%%)     |
|  :prometheus: Alert Manager                                                |                                                        | [:octicons-link-external-24: alertmanager.%%{domain}%%](https://alertmanager.%%{domain}%%)   |
| :grafana: Grafana                                                      | Monitoring Dashboards                                  | [:octicons-link-external-24: grafana.%%{domain}%% ](https://grafana.%%{domain}%%)|
| Karma | Multi-Cluster alert viewer | [:octicons-link-external-24: karma.%%{domain}%%](https://karma.%%{domain}%%)  |
| [:octicons-link-external-24: Grafana Operator](https://github.com/integr8ly/grafana-operator/) | Declarative/GitOps configuration of grafana dashboards |
| [:octicons-link-external-24: prometheus-operator](https://coreos.com/operators/prometheus/docs/latest/api.html) | Declarative management of prometheus configuration     |

`karina.yml`
```yaml
domain: %%{domain}%%
monitoring:
  prometheus:
    version: v2.20.0
    persistence:
      capacity: 20Gi
```

Deploying using :

```bash
karina deploy monitoring -c karina.yml
```

See the [user guide](/user-guide/prometheus.md) for scraping metrics and configuring grafana dashboards

For configuring long-term multi-cluster metrics see [thanos](./thanos.md)

