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
