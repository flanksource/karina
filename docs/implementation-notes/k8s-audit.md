# Implementation notes for k8s cluster auditing

## Overview

## Implementation Approach

  The auditing is *kube-api-server*-based.
  *kube-api-server* startup needs to be modified to inject a new audit-spec.

  - [ ]  Assumption  - platform-cli clusters are `kubeadm`-based
  - [x]  Verify kubeadm config of apiserver
     - [x] works on KIND
     - [ ] works in konfigadm images
  - [ ]  Find codebase for kind/kubeadm images that governs *kubeadm* configs
  - [ ]  Find injection/configuration approach


## Documentation

### Audit

* [k8s audit](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/)
* [Wehooks sampe with Falco](https://kubernetes.io/docs/tasks/debug-application-cluster/falco/)

* [Article on setup for kubeadm cluster](https://medium.com/faun/kubernetes-on-premise-cluster-auditing-eb8ff848fec4)
 
* [Blogpost for injecting into `kubeadm init`](https://evalle.xyz/posts/how-to-enable-kubernetes-auditing-with-kubeadm/)

### Kubeadm

* [Override kubeadm flags](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/control-plane-flags/)


## POC

Kind config:

**`kind-example-config.yaml`**
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
# patch the generated kubeadm config with some extra settings
kubeadmConfigPatches:
- |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    kubernetesVersion: v1.16.0
    apiServer:
      extraArgs:
        audit-policy-file: /etc/kubernetes/policies/audit-policy.yaml
        audit-log-path: /var/log/apiserver/audit.log
        audit-log-maxage: 30
        audit-log-maxsize: 200
nodes:
- role: control-plane
  image: kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62
  extraMounts:
  - hostPath: ./audit-policy.yaml
    containerPath: /etc/kubernetes/policies/audit-policy.yaml
  - hostPath: ./audit.log
    containerPath: /var/log/apiserver/audit.log
```

```bash
kind create cluster --name audit-poc --config kind-config-audit.yaml

```