First deploy the Elastic cloud on kubernetes:



`karina.yml`

```yaml
eck:
  version: 1.0.0
```

`karina deploy eck -c karina.yml`

Then create an elastic config:

`elastic-stack.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: eck
  labels:
    quack.pusher.com/enabled: "true"
---
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: logs
  namespace: eck
spec:
  version: 7.10.2
  nodeSets:
    - name: default
      count: 3
      config:
        node.master: true
        node.data: true
        node.ingest: true
        node.store.allow_mmap: false
        xpack.security.transport.ssl.supported_protocols: TLSv1.1,TLSv1.2
        xpack.security.authc.anonymous.roles: fluentd
      podTemplate:
        spec:
          containers:
            - name: elasticsearch
              env:
                - name: ES_JAVA_OPTS
                  value: -Xms4g -Xmx4g
              resources:
                requests:
                  memory: 10Gi
                  cpu: 1
                limits:
                  memory: 10Gi
                  cpu: 4
      volumeClaimTemplates:
        - metadata:
            name: elasticsearch-data
          spec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 500Gi

---
apiVersion: kibana.k8s.elastic.co/v1
kind: Kibana
metadata:
  name: logs
  namespace: eck
spec:
  version: 7.10.2
  count: 1
  elasticsearchRef:
    name: logs

---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: logs-ing
  namespace: eck
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/client_max_body_size: "256m"
    kubernetes.io/tls-acme: "true"
spec:
  tls:
    - secretName: logs-tls
      hosts:
        - logs.{{.Domain}}
  rules:
    - host: logs.{{.Domain}}
      http:
        paths:
          - backend:
              serviceName: logs-es-http
              servicePort: 9200

---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: kibana-ing
  namespace: eck
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    kubernetes.io/tls-acme: "true"
spec:
  tls:
    - secretName: kibana-tls
      hosts:
        - kibana.{{.Domain}}
  rules:
    - host: kibana.{{.Domain}}
      http:
        paths:
          - backend:
              serviceName: logs-kb-http
              servicePort: 5601

```

`kubectl deploy -f instance.yaml`