# Implementation notes for k8s cluster auditing

## Overview

## Implementation Approach

  The auditing is *kube-api-server*-based.
  *kube-api-server* startup needs to be modified to inject a new audit-spec.

  - [X]  Assumption  - platform-cli clusters are `kubeadm`-based
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

**`kind-config-audit.yaml`**
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62
    # These mounts are from KIND host to master node container
    extraMounts:
      - hostPath: ./audit-policy.yaml
        containerPath: /etc/kubernetes/policies/audit-policy.yaml
      # we're mounting the _directory_ and not just the log _file_
      # otherwise logfile rotation, etc. will fail
      - hostPath: ./auditdir/
        containerPath: /var/log/apiserver/
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
            audit-log-maxage: "30"   # quoted to prevent serialisation issues
            audit-log-maxsize: "200" # quoted to prevent serialisation issues
          # These mounts are from master node container to inner ContainerD
          # kube-apiserver container
          extraVolumes:
          - name: audit-spec
            hostPath: /etc/kubernetes/policies/audit-policy.yaml
            mountPath: /etc/kubernetes/policies/audit-policy.yaml
            readOnly: true
            pathType: File
          - name: audit-log
            hostPath: /var/log/apiserver/
            mountPath: /var/log/apiserver/
            readOnly: false
            pathType: DirectoryOrCreate

```

```bash
kind create cluster --name audit-poc --config kind-config-audit.yaml

```