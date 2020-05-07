### Configure Kubernetes Auditing 

Karina supports the kubernetes [auditing](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) using the log backend, which writes audit events to files.

Update the `kubernetes.auditing` section:

```yaml
kubernetes:
  auditing:
    policyFile: ./test/fixtures/audit-policy.yaml
  apiServerExtraArgs:
    "audit-log-path": /var/log/audit/cluster-audit.log
    "audit-log-maxsize": 1024
    "audit-log-maxage": 2
    "audit-log-maxbackup": 3
    "audit-log-format": legacy   # default is json
```

!!! warning
    Note that auditing options are only used on provisioning, to update or add auditing to an existing cluster the configuration needs to be updated and then all master nodes rolled.

For an example policy see [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml)


Relevant `apiServerExtraArgs` options:

| Key                   | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| `audit-log-path`      | The path in the api-server pod that the audit logs are written to. <br>(a value of `'-'` indicates logging to the pod logs.) <br/>If not specified, it defaults to `/var/log/audit/cluster-audit.log` |
| `audit-log-maxage`    | The maximum number of days to retain log files               |
| `audit-log-maxbackup` | The maximum number of audit log files to retain              |
| `audit-log-maxsize`   | The maximum size in megabytes of each log file               |
| `audit-log-format`    | Options are:<br/>`legacy` indicates 1-line text format for each event <br/> `json` indicates a structured json format. |

See the  [docs](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#log-backend) for a full list of options