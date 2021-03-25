
`karina.yml`
```yaml
minio:
  version: RELEASE.2020-09-02T18-19-50Z
  access_key: minio
  secret_key: minio123
  # 1 or non HA, otherwise must be in multiples of 4
  replicas: 1
```

:1: Deploy:

```bash
karina deploy minio -c karina.yml
```

:2: Access the UI: [:octicons-link-external-24: https://minio.%%{domain}%%](https://minio.%%{domain}%%)

!!! warning
    Minio has significant limitations around scaling - Once created the the number of nodes cannot be increased to scale capacity
