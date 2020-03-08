GitOps

### Persistent Volumes

### Namespace Configuration

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

See [fluent-plugin-elasticsearch](https://github.com/uken/fluent-plugin-elasticsearch) for available elastic search options

See https://github.com/vmware/kube-fluentd-operator#using-the-labels-macro

To ingest logs from a container [see](https://github.com/vmware/kube-fluentd-operator#ingest-logs-from-a-file-in-the-container)


