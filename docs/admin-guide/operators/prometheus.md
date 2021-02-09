To deploy a default monitoring stack of prometheus, alert-manager and grafana:

`karina.yml`

```yaml
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

