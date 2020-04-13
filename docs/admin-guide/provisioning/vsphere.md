## Platform Quickstart

### 1. Install platform-cli

Choose to build from source, download a release or use docker.

#### Clone,build from source and install:

```bash
git clone git@github.com:flanksource/platform-cli.git
cd platform-cli
make setup
make pack
make
make compress
sudo make install #enter password when prompted
```

_or_

#### Download release and install:

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/platform-cli/releases/latest).

Make it executable and place in your path.

_or_

#### Use docker image:

Use latest docker image:

docker pull flanksource/platform-cli:latest

## 2. Setup and verify vSphere connectivity

Make sure the following environment variables are set:

`GOVC_FQDN`
`GOVC_USER`
`GOVC_PASS`
`GOVC_DATACENTER`
`GOVC_NETWORK`

3. Create CA certs for use with the cluster

```yaml
platform-cli ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar`
```

1. Setup [environment variables](#environment-variables) and [platform configuration](#platform-configuration)
2. Download and install the platform-cli binary
3. Generate a CA for the cluster: `platform-cli ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/ingress-ca.key --password foobar`
4. Create the cluster `platform-cli provision cluster -c cluster.yaml`see [Cluster Lifecycle](#cluster-lifecycle)
5. Check the status of running vms: `platform-cli status`
6. Export an X509 based kubeconfig: `platform-cli kubeconfig admin`
7. Export an OIDC based kubeconfig: `platform-cli kubeconfig sso`
8. Build the base platform configuration: `platform-cli build all`
9. Deploy the platform configuration: `platform-cli deploy all`
10. Run conformance tests: `platform-cli test`
11. Tear down the cluster: `platform-cli cleanup`

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
# The root CA used to sign generated certs
ca:
  cert: .certs/root-ca.crt
  privateKey: .certs/root-ca.key
  password: foobar
 The VM configuration for master nodes
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
