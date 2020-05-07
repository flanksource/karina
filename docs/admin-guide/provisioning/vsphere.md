## Platform Quickstart

### 1. Install platform-cli

Choose to build from source, download a release or use docker.

#### Download release and install:

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/platform-cli/releases/latest).

Make it executable and place in your path.

_or_

#### Use docker image:

Use latest docker image:

```
docker pull flanksource/platform-cli:latest
```

## 2. Setup and verify vSphere connectivity

`platform-cli` uses the [VMWare govmomi library](https://github.com/vmware/govmomi)  which normally uses the following environment variables for connectivity parameters.

Make sure the following environment variables are set:

`GOVC_FQDN`

`GOVC_USER`

`GOVC_PASS`

`GOVC_DATACENTER`

`GOVC_NETWORK`

`GOVC_DATASTORE`

`GOVC_CLUSTER`

Connectivity can be verified by [installing govc](https://github.com/vmware/govmomi/tree/master/govc#installation) and running the following:

```bash
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"
govc about
```

## 3. Create CA certs for use with the cluster

Create a Certificate Authority for the cluster and the cluster ingress by running:

```shell
platform-cli ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar
platform-cli ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar
```

## 4. Configure the platform config

`platform-cli` uses a YAML configuration file.

Below is a small working sample.

See other examples in the [test vSphere platform fixtures](https://github.com/flanksource/platform-cli/tree/master/test/vsphere).

See the [Configuration Reference](./reference/config.md) for details of available configurations.

```yaml
##
## Sample platform config
##

## Cluster name
name: example-cluster

## Prefix to be added to VM hostnames,
hostPrefix: ex

## Endpoint for externally hosted consul cluster
## NOTE: a working consol config required to verify
##       that primary master is available.
consul: 10.100.0.13

## Domain that cluster will be available at
## NOTE: domain must be supplied for vSphere clusters
domain: 10.100.0.0.nip.io

# vSphere datacenter
datacenter: lab

dns:
  disabled: true

# The CA certs generated in step 3
ca:
  cert: .certs/root-ca.crt
  privateKey: .certs/root-ca.key
  password: foobar
ingressCA:
  cert: .certs/ingress-ca.crt
  privateKey: .certs/ingress-ca.key
  password: foobar

# A list of folders of kubernetes specs to apply, 
# these will be templatized
specs: 
  - ./manifests
versions:
  kubernetes: v1.16.4
serviceSubnet: 10.96.0.0/16
podSubnet: 10.97.0.0/16

## The VM configuration for master nodes
master:
  count: 1
  cpu: 2  #NOTE: minimum of 2
  memory: 4
  disk: 10
  network: !!env GOVC_NETWORK
  cluster: !!env GOVC_CLUSTER
  prefix: m
  template: "k8s-1.16.4"
workers:
  worker-group-a:
    prefix: w
    network: !!env GOVC_NETWORK
    cluster: !!env GOVC_CLUSTER
    count: 1
    cpu: 2
    memory: 4
    disk: 10
    template: k8s-1.16.4
```

## 5. Provision the cluster

Provision the cluster with:

```bash
platform-cli provision vsphere-cluster -c cluster.yaml
```

## 6. Deploy a CNI

Deploy Calico:

```bash
platform-cli deploy calico -c cluster.yaml
```

## 7. Deploy base configs

```bash
platform-cli deploy base -c cluster.yaml
```

## 8. Access the cluster

Export a kubeconfig file (using an X509 admin example):

```bash
platform-cli kubeconfig admin -c cluster.yaml > kubeconfig.yaml
export KUBECONFIG=$PWD/kubeconfig.yaml
```

For the session `kubectl` commands can then be used to access the cluster, e.g.:

```bash
kubectl get nodes
```

## 9. Run conformance tests

Run:

```bash
platform-cli test all -c cluster.yaml
```

## 10. Tear down the cluster

Run:

```bash
platform-cli cleanup -c cluster.yaml
```