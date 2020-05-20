Karina supports deploying the same workload on multiple clusters, in order to facilitate this dynamic ingress names are supported.

Create the ingress as usual and use `{{.Domain}}` where you would normally use the cluster wildcard DNS entry, The template will be replaced at runtime by Quack


```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: kibana-ing
  namespace: eck
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
spec:
  tls:
    - hosts:
        - kibana.{{.Domain}}
  rules:
    - host: kibana.{{.Domain}}
      http:
        paths:
          - backend:
              serviceName: logs-kb-http
              servicePort: 5601
```



### Variables available for templating

| Variable | Description                                                  |
| -------- | ------------------------------------------------------------ |
| Domain   | Wildcard domain specified at cluster provisioning and at which DNS points to the ingress layer |
| Name     | Unique name of the cluster                                   |

