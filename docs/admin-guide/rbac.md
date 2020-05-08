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