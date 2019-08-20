##  Platform

Whats included:

## Core Platform
- [x] LocalPersistent Volumes
- [x] OIDC based authentication using Dex and LDAP
- [x] Monitoring stack powered by Prometheus, Grafana and Alert Manager
- [x] Kubeadm config generation for control plane and worker nodes
- [x] Nginx based Ingress
- [x] Cloud-init generation using Konfigadm
- [x] Provisioning clusters on vSphere using govc, heavily based on CAPV
- [ ] DNS Integration via external-dns
- [ ] Backups using [velero](https://github.com/heptio/velero)
- [ ] Certificate Management using [cert-manager](https://github.com/jetstack/cert-manager)
- [ ] GitOps using [faros](https://github.com/pusher/faros)
- [ ] Per Namespace log shipping using [fluent-operator](https://github.com/vmware/kube-fluentd-operator)
- [ ] Namespace templating using [namespace-configuration-operator](https://github.com/redhat-cop/namespace-configuration-operator/tree/master)
- [ ] In cluster templating [quack](https://github.com/pusher/quack) for dynamic ingress domain names
- [ ] Authenticating proxy using [OAuth Proxy](https://github.com/pusher/oauth2_proxy) or [SSO Operator](https://github.com/jenkins-x/sso-operator)
- [ ] Conformance testing using [sonobuoy](https://github.com/heptio/sonobuoy)
- [ ] Policy Enforcement using [Gatekeeper](https://github.com/open-policy-agent/gatekeeper) (OPA)

## Management Platform
- [ ] HashiCorp Vault
- [ ] Postgres Operator
- [ ] Harbor Registry
- [ ] Multi-Cluster Log aggregation using ELK
- [ ] Multi-Cluster Metric aggregation using Thanos
- [ ] Multi-Cluster Billing using [operator-metering](- https://github.com/operator-framework/operator-metering)


## Configuration

```yaml
# DNS Wildcard domain that this cluster will be accessible under
domain:
# externally hosted consul cluster
consul:
name:
ldap:
  dn:
  server:
  adminGroup:
specs:
  - ./manifests
versions:
  kubernetes: v1.15.0
serviceSubnet: 10.96.0.0/16
podSubnet: 10.97.0.0/16
hostPrefix:
master:
  count: 5
  cpu: 4
  memory: 16
  disk: 200
  # GOVC_NETWORK
  network:
  # GOVC_CLUSTER
  cluster:
  template: T
workers:
  worker: # multiple worker groups can be specified
    count: 20
    cpu: 16
    memory: 64
    disk: 300
    # maps to vCenter parameters
    network:
    cluster:
    template:
```

Environment Variables:

```bash
export GOVC_FQDN=
export GOVC_DATACENTER=
export GOVC_CLUSTER=
export GOVC_FOLDER=
export GOVC_NETWORK=
export GOVC_PASS=
export GOVC_USER=
export GOVC_DATASTORE=
export GOVC_RESOURCE_POOL=
export GOVC_INSECURE=1
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"
```
