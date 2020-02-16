```
fluentd-config
```

### Ingress Templating

Annotate your namespace with: 

`quack.pusher.com/enabled`

### Routing Logs

https://github.com/vmware/kube-fluentd-operator 

Create a config ap in your namespace called `fluentd-config.conf` 

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <match **>
      @type elasticsearch
      host logs-es-http.eck.cluster.local
      port 9200
      index_name logs
    </match>

```

See https://github.com/uken/fluent-plugin-elasticsearch



See https://github.com/vmware/kube-fluentd-operator#using-the-labels-macro

https://github.com/vmware/kube-fluentd-operator#ingest-logs-from-a-file-in-the-container

https://github.com/vmware/kube-fluentd-operator#ingest-logs-from-a-file-in-the-container

