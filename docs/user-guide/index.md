# User Guide

| Annotation                                             | Applies To                           | Description                                                  |
| ------------------------------------------------------ | ------------------------------------ | ------------------------------------------------------------ |
| `reload/all=true`                                      | Deployments, StatefulSet, DaemonSets | Trigger redeployment if any configmap or secret is updated   |
| `reload/configmap=foo-configmap,bar-configmap`         | Deployments, StatefulSet, DaemonSets | Trigger redeployment if specified configmap(s) updated       |
| `reload/secret=foo-secret,bar-secret`                  | Deployments, StatefulSet, DaemonSets | Trigger redeployment if specified secret(s) is updated       |
| `openpolicyagent.org/webhook=ignore`                   | Namespace                            | Disable open policy validaton in this namespace              |
| `quack.pusher.com/enabled=true`                        | Namespace                            | Enable templating of ConfigMap and Ingress objects           |
| `auto-delete=1d`                                       | Namespace                            | Automatically delete a namespace after the configured timeout |
| `platform.flanksource.com/restrict-to-groups`          | Ingress                              | Restrict access to the specified Ingress to authenticated users with membership in the configured groups |
| `platform.flanksource.com/extra-configuration-snippet` | Ingress                              | Extra nginx configuration snippet to apply                   |
| `platform.flanksource.com/pass-auth-headers`           | Ingress                              | Authentication headers to pass through to the backend, a `Authentication: Bearer` header with a JWT token is sent to backends by default |

