`karina.yml`

```yaml
templateOperator:
  version: v0.1.9
```

Deploy using:

```bash
karina deploy template-operator -c karina.yml
```



### Namespace Request Example

:1: First define a template based on a CRD:

`template-definition.yml`

```yaml
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: namespacerequests.acmp.corp
spec:
  group: acmp.corp
  names:
    kind: NamespaceRequest
    listKind: NamespaceRequestList
    plural: namespacerequests
    singular: namespacerequest
  scope: Cluster
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
---
apiVersion: templating.flanksource.com/v1
kind: Template
metadata:
  name: namespacerequest
spec:
  source:
    apiVersion: acmp.corp/v1
    kind: NamespaceRequest
  resources:
    - apiVersion: v1
      kind: Namespace
      metadata:
        name: "{{.metadata.name}}"
        annotations:
          team: "{{.spec.team}}"
          service: "{{.spec.service}}"
          company: "{{.spec.company}}"
          environment: "{{.spec.environment}}"

    - apiVersion: v1
      kind: ResourceQuota
      metadata:
        name: compute-resources
        namespace: "{{.metadata.name}}"
      spec:
        hard:
          requests.cpu: "1"
          requests.memory: 10Gi
          limits.cpu: "{{ math.Div .spec.memory 8 }}m"
          limits.memory: "{{.spec.memory}}Gi"
          pods: "{{ math.Mul .spec.memory 6 }}"
          services.loadbalancers: "0"
          services.nodeports: "0"

    - apiVersion: rbac.authorization.k8s.io/v1
      kind: RoleBinding
      metadata:
        name: creator
        namespace: "{{.metadata.name}}"
      subjects:
        - kind: Group
          name: "{{.spec.team}}"
          apiGroup: rbac.authorization.k8s.io
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: namespace-admin

```

:2:
```bash
kubectl apply -f template-definition.yml
```

:3: create 1 or more namespace requests:

`request.yml`

```yaml
apiVersion: acmp.corp/v1
kind: NamespaceRequest
metadata:
  name: a
spec:
  team: blue-team
  memory: 16
```

```bash
kubectl apply -f request.yml
```



The template operator will then pick up the new *NamespaceRequest* and create the corresponding *Namespace*, ResourceQuota and *RoleBinding* objects.
