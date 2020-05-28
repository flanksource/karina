

| Annotation                                           | Description                                                  |
| ---------------------------------------------------- | ------------------------------------------------------------ |
| `nginx.ingress.kubernetes.io/ssl-redirect=false`     | Prevent automatic redirect from HTTP to HTTPS                |
| `nginx.ingress.kubernetes.io/backend-protocol=HTTPS` | Use a backend protocol other than HTTP to connect to upstream services. Can be `HTTPS`, `GRPC`, `GRPCS` and `AJP` |
| `kubernetes.io/tls-acme=true`                        | Automatically generate and sign a new certificate            |



For a full list of supported nginx annotation see [here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/)



#### Automatic Certificate Generation

```yaml
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: kibana-ing
  annotations:
    # Turn on automatic cert generation for this ingress
    kubernetes.io/tls-acme: "true"
spec:
  tls:
    # Must specify a secretName to store the cert, it does not need to exist.
    - secretName: kibana-tls
      hosts:
        - kibana.{{.Domain}}

```



#### Specifying your own certificate request details

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: example-com
  namespace: sandbox
spec:
  # Specify this secretName in your ingress
  secretName: example-com-tls
  duration: 2160h # 90d
  renewBefore: 360h # 15d
  organization:
  - jetstack
  # The use of the common name field has been deprecated since 2000 and is
  # discouraged from being used.
  commonName: example.com
  isCA: false
  keySize: 2048
  keyAlgorithm: rsa
  keyEncoding: pkcs1
  usages:
    - server auth
    - client auth
  dnsNames:
  - example.com
  - www.example.com
  issuerRef:
    # ingress-issuer is created by default, but you can specify any CertManager issuer available on the cluster
    name: ingress-issuer
    kind: ClusterIssuer
```

For a full description of the certificate options see [cert-manager](https://cert-manager.io/docs/usage/certificate/)

#### Dynamic Ingress Hostname

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

