

| Annotation                                             | Description                                                  |
| ------------------------------------------------------ | ------------------------------------------------------------ |
| `nginx.ingress.kubernetes.io/ssl-redirect=false`       | Prevent automatic redirect from HTTP to HTTPS                |
| `nginx.ingress.kubernetes.io/backend-protocol=HTTPS`   | Use a backend protocol other than HTTP to connect to upstream services. Can be `HTTPS`, `GRPC`, `GRPCS` and `AJP` |
| `kubernetes.io/tls-acme=true`                          | Automatically generate and sign a new certificate            |
| `platform.flanksource.com/restrict-to-groups`          | Restrict access to the specified Ingress to authenticated users with membership in the configured groups |
| `platform.flanksource.com/extra-configuration-snippet` | Extra nginx configuration snippet to apply                   |
| `platform.flanksource.com/pass-auth-headers`           | Authentication headers to pass through to the backend, a `Authentication: Bearer` header with a JWT token is sent to backends by default |

For a full list of supported nginx annotation see [here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/)



### Ingress Authentication

Using a combination of Dex and [Oauth2-Proxy](https://github.com/oauth2-proxy/oauth2-proxy) you can configure ingress'es to require authentication and membership in specific groups:


### Dynamic Ingress Hostname

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

