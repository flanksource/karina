
#### Automatic Certificate Generation

```yaml
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


For a full description of the certificate options see [cert-manager](https://cert-manager.io/docs/usage/certificate/)

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
