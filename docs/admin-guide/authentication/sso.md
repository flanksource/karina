# Setting up SSO

### Configure LDAP in `config.yaml`.

```yaml
ldap:
  adminGroup: NA1
  username: uid=admin,ou=system
  password: secret
  port: 10636
  host: apacheds.ldap
  dn: ou=users,dc=example,dc=com
```

### Deploy ApacheDS and Dex

```shell
$ platform-cli deploy stubs
$ platform-cli deploy dex
```

### Wait for pods to start:

```
$ kubectl -n ldap get po
NAME                        READY   STATUS    RESTARTS   AGE
apacheds-56c656465d-ttgv7   1/1     Running   0          1m

$ kubectl -n dex get po
NAME                   READY   STATUS    RESTARTS   AGE
dex-56dd8bff8f-59vkc   1/1     Running   0          1m
dex-56dd8bff8f-kt8rt   1/1     Running   1          1m
dex-56dd8bff8f-l4vqg   1/1     Running   0          1m
```

### Setup RBAC permissions for users

```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-rbac-ldap
rules:
  - apiGroups: ["*"]
    resources: ["pods", "nodes"]
    verbs: ["list"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-rbac-role
subjects:
  - apiGroup: ""
    kind: User
    name: test@example.com
roleRef:
  apiGroup: ""
  kind: ClusterRole
  name: test-rbac-ldap
```
