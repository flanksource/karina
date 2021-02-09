# Kind Quickstart

### 1) Install karina

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/karina/releases/latest).

#### Linux
```bash
wget -nv -nc -O karina https://github.com/flanksource/karina/releases/latest/download/karina && chmod +x karina
mv karina /usr/local/bin/karina
```

#### MacOSX
```bash
wget -nv -nc -O karina https://github.com/flanksource/karina/releases/latest/download/karina_osx && chmod +x karina
mv karina_osx /usr/local/bin/karina
```

!!! info
    For production pipelines you should always pin the version of karina you are using

### 2) Create a configuration file:

`test-cluster.yml`

```yaml
domain: 127.0.0.1.nip.io
name: test-cluster
calico:
  version: v3.8.2
kubernetes:
  version: v1.18.15
  masterIP: localhost
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

### 3) Generate the necessary CA's:

```shell
# generate CA for kubernetes api-server authentication
karina ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1

# generate ingressCA for ingress certificates
karina ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1

```

### 4) Provision the cluster in kind:

```shell
karina provision kind-cluster -c test-cluster.yml
```

### 5) Deploy the bare minimum configuration:
This command will deploy following:
- CRDs for all production runtime defined in the 

```shell
karina deploy phases --crds --base --dex --calico -c test-cluster.yml
```

### 6) Deploy everything else that may be configured:

```shell
karina deploy all
```

### Cleanup
Stop and delete the container running Kind with
```shell
kind delete cluster --name test-cluster-control-plane
```


## Troubleshooting

KIND cluster creation issues can be debugged by specifying the `--trace` argument to `karina` during creation:

```shell
karina provision kind-cluster --trace
```
