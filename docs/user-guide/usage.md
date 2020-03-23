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

### Configmap reloader

https://github.com/stakater/Reloader

Reloader will trigger pod re-creation when used ConfigMaps or Secrets are being changed.

To recreate pods on configmap/secret change add this annotations to Deployment, DaemonSet or StatefulSet objects:

For ConfigMap:
```
configmap.reloader.stakater.com/reload: "configmap-name"
```

For Secret:
```
secret.reloader.stakater.com/reload: "secret-name"
```

You can also specify multiple names with comma and you can use this annotation to reload on all configmaps and secret changes:

```
reloader.stakater.com/auto: "true"
```
