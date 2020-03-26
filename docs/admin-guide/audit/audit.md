# Cluster Auditing

### Configure Auditing in `config.yaml`.

```yaml
auditConfig:
  #disabled: true
  auditPolicyFile: ./test/fixtures/audit-policy.yaml
  kubeApiServerOptions:
    # Audit Log configs:
    # audit-log-path: /var/auditLogs/audit.logs
    audit-log-path: "-" # log to stdout, i.e. apiserver logs
    audit-log-maxage: 2
    audit-log-maxbackup: 3
    audit-log-maxsize: 10
    audit-log-format: json
    # Audit Webhook configs:
    audit-webhook-config-file: ./test/fixtures/audit-webhoop-config.yaml
    audit-webhook-initial-backoff: 6
```

Implicit default is:

```yaml
auditConfig:
  disabled: true
```