## Shipping

???+ asterix "Prerequisites"
     A cluster-wide [logging](/admin-guide/logging) collector has been setup by the cluster administrator


To ingest your applications logs your pod needs to be annotated with:  `%%{ filebeat.prefix }%%/enabled: true`, This can be done either at the Namespace level or inside your workloads `.spec.template.metadata.annotations` fields.


### Filtering

To filter overly verbose messages from being collected, add one or more <a href="https://www.elastic.co/guide/en/beats/filebeat/7.10/drop-event.html" target=_blank>:material-open-in-new: drop_event</a>


```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
         %%{ filebeat.prefix }%%/processors.0.drop_event.when.contains.message: DEBUG
         %%{ filebeat.prefix }%%/processors.1.drop_event.when.contains.message: TRACE
```


### Parsing
To tokenize unstructured messages into structured fields before use the <a href="https://www.elastic.co/guide/en/beats/filebeat/7.10/dissect.html" target=_blank>:material-open-in-new: dissect</a>


Given the following log line:
```
2021-02-14 07:35:46.222  INFO 1 --- [           main] o.h.h.i.QueryTranslatorFactoryInitiator  : HHH000397: Using ASTQueryTranslatorFactory
```

Using a tokenization string of: `%{date} %{time}  %{level} %{} %{} [%{entry}] %{class}: %{message}` will parse it into:


```json
{
  "class": "o.h.h.i.QueryTranslatorFactoryInitiator  ",
  "date": "2021-02-14",
  "entry": "           main",
  "level": "",
  "message": "HHH000397: Using ASTQueryTranslatorFactory",
  "time": "07:35:46.222"
}
```

!!! tip
    <a href="https://dissect-tester.jorgelbg.me" target="_blank">:material-open-in-new: dissect-tester</a> can help testing tokenization strings with sample logs

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
         %%{ filebeat.prefix }%%/processors.0.dissect.tokenizer: '%{date} %{time}  %{level} %{} %{} [%{entry}] %{class}: %{message}'
         %%{ filebeat.prefix }%%/processors.0.dissect.ignore_failure: "true"
         %%{ filebeat.prefix }%%/processors.0.dissect.target_prefix: ""
         %%{ filebeat.prefix }%%/processors.0.dissect.overwrite_keys: "true"
```



### JSON

JSON formatted logs can be decoded using <a href="https://www.elastic.co/guide/en/beats/filebeat/7.10/decode-json-fields.html" target=_blank>:material-open-in-new: decode-json-fields</a>


```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
         %%{ filebeat.prefix }%%/processors.0.decode_json_fields.fields.0: message
         %%{ filebeat.prefix }%%/processors.0.decode_json_fields.target: ""
         %%{ filebeat.prefix }%%/processors.0.decode_json_fields.overwrite_keys: "true"
         %%{ filebeat.prefix }%%/processors.0.decode_json_fields.add_error_key: "true"
```

###  Multiline

Multi-line log messages such as Java stack traces can be combined using <a href="https://www.elastic.co/guide/en/beats/filebeat/7.10/multiline-examples.html" target=_blank>:material-open-in-new: multiline</a>

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
        %%{ filebeat.prefix }%%/multiline.pattern: "^[[:space:]]+(at|\.{3})[[:space:]]+\b|^Caused by:"
        %%{ filebeat.prefix }%%/multiline.negate: "true"
        %%{ filebeat.prefix }%%/multiline.match: after
```


### Example

```yaml
[[% include './user-guide/petclinic.yaml' %]]
```



## Realtime Tailing

???+ asterix "Prerequisites"
     [stern](/user-guide/install#install-stern) is installed

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

