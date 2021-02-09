

`config.yaml`

```yaml
filebeat:
  - name: infra
    version: 7.10.2
    index: filebeat-infra
    prefix: com.flanksource.infra
    elasticsearch:
      url: logs-es-http.eck.svc.cluster.local
      user: elastic
      password: elastic
      port: 9200
      scheme: https
journalbeat:
  version: 7.10.2
  elasticsearch:
    url: logs-es-http.eck.svc.cluster.local
    user: elastic
    password: elastic
    port: 9200
    scheme: https
auditbeat:
  disabled: true
  version: 7.10.2
  elasticsearch:
    url: logs-es-http.eck.svc.cluster.local
    user: elastic
    password: elastic
    port: 9200
    scheme: http
packetbeat:
  version: 7.10.2
  elasticsearch:
    url: logs.127.0.0.1.nip.io
    user: elastic
    password: elastic
    port: 443
    scheme: https
  kibana:
    url: kibana.127.0.0.1.nip.io
    user: elastic
    password: elastic
    port: 443
    scheme: https
```

`karina deploy phases --packetbeat --auditbeat --journalbeat --filebeat`



See [User Guide -> Logging](../user-guide/logging) for details on how to configure logging per-namespace or per-pod.