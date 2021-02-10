???+ asterix "Prerequisites"
     [karina](/admin-guide/#installing-karina) is installed



## Generate CA's

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


## Create karina.yml

`karina.yml`
```yaml
domain: 127.0.0.1.nip.io
name: test-cluster
calico:
  version: v3.8.2
kubernetes:
  version: %%{ kubernetes.version }%%
  containerRuntime: containerd
ca:
  cert: .certs/root-ca.crt
  privateKey: .certs/root-ca.key
  password: foobar
ingressCA:
  cert: .certs/ingress-ca.crt
  privateKey: .certs/ingress-ca.key
  password: foobar
```
## Provision the cluster

```shell
karina provision kind-cluster -c karina.yml
```

## Deploy the bare config
```shell
karina deploy phases --crds --base --dex --calico -c karina.yml
```

## Deploy everything else

```shell
karina deploy all -c karina.yml
```

## Cleanup
Stop and delete the container running Kind with
```shell
kind delete cluster --name test-cluster-control-plane
```


## Troubleshooting

Kind cluster creation issues can be debugged by specifying the `--trace` argument to `karina` during creation:

```shell
karina provision kind-cluster --trace
```
