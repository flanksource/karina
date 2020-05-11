# User Guide

| Annotation                                     | Applies To                           | Description                                                |
| ---------------------------------------------- | ------------------------------------ | ---------------------------------------------------------- |
| `reload/all=true`                              | Deployments, StatefulSet, DaemonSets | Trigger redeployment if any configmap or secret is updated |
| `reload/configmap=foo-configmap,bar-configmap` | Deployments, StatefulSet, DaemonSets | Trigger redeployment if specified configmap(s) updated     |
| `reload/secret=foo-secret,bar-secret`          | Deployments, StatefulSet, DaemonSets | Trigger redeployment if specified secret(s) is updated     |
| `openpolicyagent.org/webhook=ignore`           | Namespace                            | Disable open policy validaton in this namespace            |
| `quack.pusher.com/enabled=true`                | Namespace                            | Enable templating of ConfigMap and Ingress objects         |

