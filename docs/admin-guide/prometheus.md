

The monitoring stack includes the following components, which are all accessible via default ingress addresses

| Component                                                    | Use                                                    | Access                         |
| ------------------------------------------------------------ | ------------------------------------------------------ | ------------------------------ |
| Prometheus                                                   |                                                        | https://prometheus.DOMAIN      |
| Alert Manager                                                |                                                        | https://alertmanager.DOMAIN    |
| [prometheus-operator](https://coreos.com/operators/prometheus/docs/latest/api.html) | Declarative management of prometheus configuration     |                                |
| Grafana                                                      | Monitoring Dashboards                                  | https://grafana.CLUSTER_DOMAIN |
| [Grafana Operator](https://github.com/integr8ly/grafana-operator/) | Declarative/GitOps configuration of grafana dashboards |                                |
|                                                              |                                                        |                                |

`karina.yml`
```yaml
domain: DOMAIN
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

See the [user guide ](../user-guide/monitoring.md)for scraping metrics and configuring grafana dashboards

For configuring long-term multi-cluster metrics see [thanos](./thanos.md)

