???+ asterix "Prerequisites"
    * [karina](/admin-guide/#install-karina) is installed
    * [Machine Image](./machine-images.md) with the matching versions of `kubeadm`, `kubectl` and `kubelet` and either docker or containerd
    * [Access](./vcenter.md) to a vCenter server



## :1: Generate CA's

```shell
# generate CA for kubernetes api-server authentication
karina ca generate --name root-ca \
  --cert-path .certs/root-ca.crt \
  --private-key-path .certs/root-ca.key \
  --password foobar --expiry 10

# generate ingressCA for ingress certificates
karina ca generate --name ingress-ca \
  --cert-path .certs/ingress-ca.crt \
  --private-key-path .certs/ingress-ca.key \
  --password foobar  --expiry 10
```


## :2: Create karina.yaml


`karina.yaml`
```yaml
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
  kubernetes: %%{ kubernetes.version }%%
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
  template: "kube-%%{ kubernetes.version }%%"
workers:
  worker-group-a:
    prefix: w
    network: !!env GOVC_NETWORK
    cluster: !!env GOVC_CLUSTER
    count: 1
    cpu: 2
    memory: 4
    disk: 10
    template: kube-%%{ kubernetes.version }%%
```


See other examples in the [test vSphere platform fixtures](https://github.com/flanksource/karina/tree/master/test/vsphere).

See the [Configuration Reference](config.md) for details of available configurations.


## :3: Provision the cluster

Provision the cluster with:

```bash
karina provision vsphere-cluster -c karina.yml
karina deploy phases --crd --base --calico -c karina.yml
```

## :4: Access the cluster

Export a kubeconfig:

```bash
karina kubeconfig admin -c karina.yml > kubeconfig.yaml
export KUBECONFIG=$PWD/kubeconfig.yaml
```

For the session `kubectl` commands can then be used to access the cluster, e.g.:

```bash
kubectl get nodes
```

## :5: Run E2E Tests


```bash
karina test all --e2e -c karina.yml
```

## :6: Tear down the cluster


```bash
karina terminate -c karina.yml
```
