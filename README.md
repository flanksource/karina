##  Platform

[![CircleCI](https://circleci.com/gh/flanksource/platform-cli.svg?style=svg)](https://circleci.com/gh/flanksource/platform-cli)

Whats included:

### Core Platform / Data Plane
- [x] Kubeadm config generation for control plane and worker nodes
- [x] [Nginx]((https://github.com/kubernetes/ingress-nginx)) based ingress
- [ ] [Contour](https://github.com/projectcontour/contour) based ingress
- [x] Calico CNI
- [ ] NSX-T CNI
- [ ] Istio
- [ ] Multi-Cluster Calico BGP Peering
- [ ] Multi-Cluster Load Balancing using [Gimbal](https://github.com/vmware-tanzu/gimbal)
- [ ] Multi-Cluster Load Balancing using NSX-T

### Control Plane

- [x] Master Service Discvovery using [consul](https://github.com/hashicorp/consul-template)
- [x] LocalPersistent Volumes using [local-path-provisioner](https://github.com/rancher/local-path-provisioner)
- [x] OIDC based authentication using [Dex](https://github.com/dexidp/dex) and LDAP
- [x] Monitoring stack powered by [Prometheus, Grafana and Alert Manager](https://github.com/coreos/kube-prometheus)
- [x] Cloud-init generation using [konfigadm](https://github.com/moshloop/konfigadm)
- [x] Provisioning clusters on vSphere using [govmomi](https://github.com/vmware/govmomi), heavily based on [CAPV](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere)
- [x] Namespace templating using [namespace-configuration-operator](https://github.com/redhat-cop/namespace-configuration-operator/tree/master)
- [ ] In cluster templating [quack](https://github.com/pusher/quack) for dynamic ingress domain names
- [ ] Per Namespace log shipping using [fluent-operator](https://github.com/vmware/kube-fluentd-operator)
- [ ] DNS Integration via [external-dns](https://github.com/kubernetes-incubator/external-dns)
- [ ] Backups using [velero](https://github.com/heptio/velero)
- [x] Certificate Management using [cert-manager](https://github.com/jetstack/cert-manager)
- [ ] GitOps using [faros](https://github.com/pusher/faros) or [flux](https://fluxcd.io)
- [ ] Authenticating proxy using [OAuth Proxy](https://github.com/pusher/oauth2_proxy) or [SSO Operator](https://github.com/jenkins-x/sso-operator)
- [ ] Conformance testing using [sonobuoy](https://github.com/heptio/sonobuoy)
- [x] Policy Enforcement using [OPA](https://github.com/open-policy-agent/kube-mgmt) (OPA)
  - [ ] Don't allow overlapping ingress domains
  - [ ] Only allow the use of whitelisted ingress domain names
  - [ ] Enforce mem / cpu requests / limits on all pods
  - [ ] Enforce quotas on all namespaces
  - [ ] Don't allow latest images
  - [ ] Force use of private registry
  - [ ] Pod Security Policies
  - [ ] Enforce liveness / health probes
  - [ ] Enforce team-labels


#### Management Platform
- [x] [Postgres Operator](https://github.com/CrunchyData/postgres-operator)
- [x] Harbor Registry
- [ ] HashiCorp Vault
- [ ] Redis Operator
- [x] Multi-Cluster Log aggregation using [ELK/ECK](https://github.com/elastic/cloud-on-k8s)
- [ ] Multi-Cluster Metric aggregation using [Thanos](https://github.com/thanos-io/thanos)
- [ ] Multi-Cluster Billing using [operator-metering](https://github.com/operator-framework/operator-metering)

### Prerequisites

* vCenter Cluster and user with admin rights on Datastore, Network and Folder
  * Populate the environment file and source it
* External Consul Cluster for master service discovery
  * Configure the `consul` field in the config
* External DNS for wildcard based ingress into namespaces
  * Once deployed, add a wildcard record in the format `*.<domain>` pointing to the F5 VIP and configure `domain` in the config
* External F5 load balancer to load balance external traffic into the cluster (DNS A records points to a F5 VIP)
  * Add all worker node IP's to the F5 member pool
* Active Directory server with a user with rights to list
  * Configure the `ldap` fields in the config

### Image Lifecycle

##### Building a new image

## Platform Quickstart

1. Setup [environment variables](#environment-variables) and [platform configuration](#platform-configuration)
2. Download and install the platform-cli binary
3. Create the cluster `platform-cli provision cluster -c cluster.yml`see [Cluster Lifecycle](#cluster-lifecycle)
4. Check the status of running vms: `platform-cli status`
5. Export an X509 based kubeconfig: `platform-cli kubeconfig admin`
6. Export an OIDC based kubeconfig: `platform-cli kubeconfig sso`
7. Build the base platform configuration: `platform-cli build all`
8. Deploy the platform configuration: `platform-cli deploy all`
9. Run conformance tests: `platform-cli test`
10. Tear down the cluster: `platform-cli cleanup`

#### PlatformConfiguration

```yaml
# DNS Wildcard domain that this cluster will be accessible under
domain:
# Endpoint for externally hosted consul cluster
consul:
# Cluster name
name:
ldap:
  # Domain binding, e.g. DC=local,DC=corp
  dn:
  # LDAPS hostname / IP
  host:
  # LDAP group name that will be granted cluster-admin
  adminGroup:
specs: # A list of folders of kubernetes specs to apply, these will be templatized
  - ./manifests
versions:
  kubernetes: v1.15.0
serviceSubnet: 10.96.0.0/16
podSubnet: 10.97.0.0/16
# Prefix to be added to VM hostnames,
hostPrefix:
# The VM configuration for master nodes
master:
  count: 5
  cpu: 4
  memory: 16
  disk: 200
  # GOVC_NETWORK
  network:
  # GOVC_CLUSTER
  cluster:
  template:
# The VM configuration for worker nodes, multiple groups can be specified
workers:
  worker:
    count: 8
    cpu: 16
    memory: 64
    disk: 300
    # GOVC_NETWORK
    network:
    # GOVC_CLUSTER
 	  cluster:
    template:
```

The PlatformConfiguration is used to generate other files used to bootstrap a cluster:

* [ClusterConfiguration](https://godoc.org/k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2#ClusterConfiguration)
* JoinConfiguration
* consul.json

##### Environment Variables

The following variables are required to configure access to vSphere, they also work with [govc](https://github.com/vmware/govmomi/tree/master/govc)

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

### Cluster Lifecycle

![](docs/Cluster%20Lifecycle.png)

##### Control Plane Init

The control plane is initialized by:

- Generating all the certificates (to be used to generate access credentials or provision new control plane nodes)
- Injecting the certificates and configs into a cloud-config file and a provisioning a VM, on boot `kubeadm init` is run to bootstrap the cluster

##### Adding Secondary Control Plane Nodes

- The certificates are injected into cloud-init and multiple VM's are provisioned concurrently which run `kubeadm join --control-plane` on boot

##### Adding Workers

- Workers have a bootstrap token injected into cloud-init and multiple VM's are provisioned concurrently which run `kubeadm --join` on boot
