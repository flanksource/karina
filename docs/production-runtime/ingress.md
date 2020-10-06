

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

#### 

See [User Guide --> Ingress](../user-guide/ingress) for more details on configuring ingress objects.

