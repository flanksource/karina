## Platform Quickstart



## vSphere Requirements

* A vcenter user and password with enough permissions to create VM's, etc..
* A ResourcePool and folder to place VM's
* A VM/Template available with the matching versions of `kubeadm`, `kubectl` and `kubelet` preinstalled
* A service discovery mechanism 


### 1. Install karina

Choose to build from source, download a release or use docker.

#### Download release and install:

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/karina/releases/latest).

Make it executable and place in your path.

_or_

#### Use docker image:

Use latest docker image:

```
docker pull flanksource/karina:v0.15.0
```

## 2. Setup and verify vSphere connectivity

Configure your environment with `GOVC_*` variables and then map them into the configuration file using the `!!env` YAML tag:

```yaml
vsphere:
  username:  !!env GOVC_USER
  datacenter: !!env GOVC_DATACENTER
  cluster: !!env GOVC_CLUSTER
  folder: !!env GOVC_FOLDER
  datastore: !!env GOVC_DATASTORE
  # can be found on the Datastore summary page
  datastoreUrl: !!env GOVC_DATASTORE_URL
  password: !!env GOVC_PASS
  hostname: !!env GOVC_FQDN
```

Connectivity can then be verified by [installing govc](https://github.com/vmware/govmomi/tree/master/govc#installation) and running the following:

```bash
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"
govc about
```

## 3. Create CA certs for use with the cluster

Create a Certificate Authority for the cluster and the cluster ingress by running:

```shell
# generate CA for kubernetes api-server authentication
karina ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1

# generate ingressCA for ingress certificates
karina ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar
```

## 4. Setup Service discovery

Karina requires a service discovery mechanism to facilitate the initial connection to the kubernetes hosts.  A containerised consul service discovery can be enabled on a host in the vsphere cluster using the [konfigadm](https://github.com/flanksource/konfigadm) tool:

```bash
konfigadm apply - << EOF
commands:
  - mkdir -p /opt/consul
  - chown -R 100:1000 /opt/consul
  - iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-ports 8500
container_runtime:
  type: docker
containers:
  - image: docker.io/consul:1.9.1
    docker_opts: --net=host
    args: agent -server -ui -data-dir /opt/consul -datacenter lab -bootstrap
    volumes:
      - /opt/consul:/opt/consul
    env:
      CONSUL_BIND_INTERFACE: ens160
      CONSUL_CLIENT_INTERFACE: ens160
EOF
```

## 5. Generate Template image

To create a template image with the prerequisite configuration installed use the [konfigadm](https://github.com/flanksource/konfigadm) tool:

```bash
# Create config file
echo >> k8s-1.18.yaml << EOF
kubernetes:
  version: 1.18.6
commands:
  - systemctl start docker
  - kubeadm config images pull  --kubernetes-version 1.18.6
  - apt remove -y unattended-upgrades
  - apt-get update
  - apt-get upgrade -y
container_runtime:
  type: docker
cleanup: true
EOF

# Build image
konfigadm images build -v --image ubuntu1804 \
   --resize +15g \
   --output-filename kube-v1.18.6.img  \
   k8s-1.18.yml

# Upload image as template (Note konfigadm expects $GOVC env vars to be present)
export NAME=k8s-v1.18.6-template
echo Pushing image to $GOVC_FQDN/$GOVC_CLUSTER/$GOVC_DATASTORE, net=$GOVC_NETWORK
konfigadm images upload ova -vv --image kube-v1.18.6.img --name $NAME

# Configure Template
govc device.serial.add -vm $NAME -
govc vm.change -vm $NAME -nested-hv-enabled=true -vpmc-enabled=true
govc vm.upgrade -version=15 -vm $NAME
```

## 6. Configure the platform config

`karina` uses a YAML configuration file.

Below is a small working sample.

See other examples in the [test vSphere platform fixtures](https://github.com/flanksource/karina/tree/master/test/vsphere).

See the [Configuration Reference](./reference/config.md) for details of available configurations.

`cluster.yaml`

```yaml
##
## Sample platform config
##

## Cluster name
name: example-cluster

## Prefix to be added to VM hostnames
hostPrefix: vsphere-k8s-

vsphere:
  username:  !!env GOVC_USER
  datacenter: !!env GOVC_DATACENTER
  cluster: !!env GOVC_CLUSTER
  folder: !!env GOVC_FOLDER
  datastore: !!env GOVC_DATASTORE
  # can be found on the Datastore summary page
  datastoreUrl: !!env GOVC_DATASTORE_URL
  password: !!env GOVC_PASS
  hostname: !!env GOVC_FQDN
  resourcePool: !!env GOVC_RESOURCE_POOL
  csiVersion: v2.0.0
  cpiVersion: v1.1.0

## Endpoint for externally hosted consul cluster
## NOTE: a working consul config required to verify
##       that primary master is available.
consul: 10.100.0.13

## Domain that cluster will be available at
## NOTE: domain must be supplied for vSphere clusters
domain: 10.100.0.0.nip.io

# Name of consul datacenter
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

versions:
  kubernetes: v1.16.4
serviceSubnet: 10.96.0.0/16
podSubnet: 10.97.0.0/16
calico:
  version: v3.8.2

## The VM configuration for master nodes
master:
  count: 1
  cpu: 2  #NOTE: minimum of 2
  memory: 4
  disk: 20
  network: !!env GOVC_NETWORK
  cluster: !!env GOVC_CLUSTER
  prefix: m
  template: k8s-v1.18.6-template
workers:
  worker-group-a:
    prefix: w
    network: !!env GOVC_NETWORK
    cluster: !!env GOVC_CLUSTER
    count: 1
    cpu: 2
    memory: 4
    disk: 20
    template: k8s-v1.18.6-template
```

## 7. Provision the cluster

Provision the cluster with:

```bash
karina provision vsphere-cluster -c cluster.yaml
```

## 8. Upload CRDs to cluster

Add the CRDs required for cluster configuration with:

```bash
karina deploy crds -c cluster.yaml
```

## 9. Deploy a CNI

Deploy Calico:

```bash
karina deploy calico -c cluster.yaml
```

## 10. Deploy base configs

```bash
karina deploy base -c cluster.yaml
```

## 11. Access the cluster

Export a kubeconfig file (using an X509 admin example):

```bash
karina kubeconfig admin -c cluster.yaml > kubeconfig.yaml
export KUBECONFIG=$PWD/kubeconfig.yaml
```

For the session `kubectl` commands can then be used to access the cluster, e.g.:

```bash
kubectl get nodes
```

## 12. Run E2E Tests

Run:

```bash
karina test all --e2e -c cluster.yaml
```

## 13. Tear down the cluster

Run:

```bash
karina terminate -c cluster.yaml
```
