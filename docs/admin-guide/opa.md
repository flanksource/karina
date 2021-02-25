#OPA Gatekeeper
## Deploying

The minimum required configuration for karina to install the gatekeeper operator is a version specification:

`karina.yml`
```yaml
gatekeeper:
  version: v3.3.0
```

---
!!! Warning Deprecated
Vanilla [:octicons-link-external-24: OPA with kube-mgmt](https://www.openpolicyagent.org/docs/kubernetes-admission-control.html) can still be deployed using the [opa](/reference/config/#opa) config flag. It is however no longer recommended or supported.
---

Deploy using :

```bash
karina deploy opa -c karina.yml

```

See the [:octicons-link-external-24: Gatekeeper Documentation](https://open-policy-agent.github.io/gatekeeper/website/docs/howto/) for general gatekeeper information.

By default, karina deploys gatekeeper with a selection of the default constrainttemplates found in the [:octicons-link-external-24: gatekeeper example library](https://github.com/open-policy-agent/gatekeeper-library).  These include:

- K8sAllowedRepos 
- K8sBannedImageTags
- K8sBlockNodePort
- K8sContainerLimits
- K8sContainerRatios
- K8sDenyAll
- K8sDisallowedTags
- K8sRequiredAnnotations
- K8sRequiredLabels
- K8sRequiredProbes
- K8sUniqueIngressHost

Additional templates and constraints can be referred to in the deployment configuration using the `constraint` and `template` fields to indicate the folders they are located in:
`karina.yml`
```yaml
gatekeeper:
  version: v3.3.0
  constraints: /path/to/constraints/folder
  templates: /path/to/templates/folder
```

## Testing constraints

To test new constraints, it is recommended to initially configure them with dryrun enforcement

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sAllowedRepos
metadata:
  name: repo-is-docker
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    namespaces:
      - "default"
  parameters:
    repos:
      - "docker.io"
  enforcementAction: dryrun
```

and monitor the number of violations that would be enforced using:

```bash
# high level overview
karina status violations -c gatekeeper.yaml
# specific violation
kubectl describe k8sallowedrepos.constraints.gatekeeper.sh/repo-is-docker
```

See [:octicons-link-external-24:  Rego Playground](https://play.openpolicyagent.org) for a useful rego testing and debugging tool

## Whitelisting Namespaces

By default, karina excludes the following namespaces from Gatekeeper policing:
 - cert-manager
 - dex
 - eck
 - elastic-system
 - gatekeeper-system
 - harbor
 - ingress-nginx
 - kube-system
 - local-path-storage
 - minio
 - monitoring
 - nsx-system
 - opa
 - platform-system
 - postgres-operator
 - quack
 - sealed-secrets
 - tekton
 - vault
 - velero
 
 Additional namespaces can be excluded at deployment using the 'whitelistNamespace' config option:
 
 ```yaml
gatekeeper:
  version: v3.3.0
  constraints: /path/to/constraints/folder
  templates: /path/to/templates/folder
  whitelistNamespaces:
    - unpolicied-namespace 
    - additional-unpolicied-namespace
```

