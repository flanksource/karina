### Ingesting logs via Elastic

| Annotation                                                   | Description                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| `co.elastic.logs/enabled`                                    | Filebeat gets logs from all containers by default, you can set this hint to false to ignore the output of the container. Filebeat won’t read or send logs from it. If default config is disabled, you can use this annotation to enable log retrieval only for containers with this set to true. |
| `co.elastic.logs/multiline.*` <br>`co.elastic.logs/multiline.pattern: '^\['` <br>`co.elastic.logs/multiline.negate: true`<br/>`co.elastic.logs/multiline.match: after` | Multiline settings. See Multiline messages for a full list of all supported options. |
|                                                              |                                                              |
| `co.elastic.logs/json.*`                                     | JSON settings. See json for a full list of all supported options. |
| `co.elastic.logs/include_lines`                              | A list of regular expressions to match the lines that you want Filebeat to include. See Inputs for more info. |
| `co.elastic.logs/exclude_lines`                              | A list of regular expressions to match the lines that you want Filebeat to exclude. See Inputs for more info. |
| `co.elastic.logs/module`                                     | Instead of using raw docker input, specifies the module to use to parse logs from the container. See Modules for the list of supported modules. |
| `co.elastic.logs/fileset`                                    | When module is configured, map container logs to module filesets. You can either configure a single fileset like this: |
| `co.elastic.logs/processors`                                 | Define a processor to be added to the Filebeat input/module configuration. See Processors for the list of supported processors |
| `co.elastic.logs/processors.1.dissect.tokenizer: "%{key1} %{key2}" co.elastic.logs/processors.dissect.tokenizer: "%{key2} %{key1}"` | In order to provide ordering of the processor definition, numbers can be provided. If not, the hints builder will do arbitrary ordering: <br> In this sample the processor definition tagged with 1 would be executed first. |

### Querying logs using stern

???+ asterix "Prerequisites"
     [stern](/user-guide/#install-stern) is installed



Tail the `gateway` container running inside of the `envvars` pod on staging
```
stern envvars --context staging --container gateway
```

Tail the `staging` namespace excluding logs from `istio-proxy` container
```
stern -n staging --exclude-container istio-proxy .
```

Show auth activity from 15min ago with timestamps
```
stern auth -t --since 15m
```

Follow the development of `some-new-feature` in minikube
```
stern some-new-feature --context minikube
```

View pods from another namespace
```
stern kubernetes-dashboard --namespace kube-system
```

Tail the pods filtered by `run=nginx` label selector across all namespaces
```
stern --all-namespaces -l run=nginx
```

Follow the `frontend` pods in canary release
```
stern frontend --selector release=canary
```

Pipe the log message to jq:
```
stern backend -o json | jq .
```

Only output the log message itself:
```
stern backend -o raw
```

Output using a custom template:

```
stern --template '{{.Message}} ({{.Namespace}}/{{.PodName}}/{{.ContainerName}})' backend
```

Output using a custom template with stern-provided colors:

```
stern --template '{{.Message}} ({{.Namespace}}/{{color .PodColor .PodName}}/{{color .ContainerColor .ContainerName}})' backend
```
