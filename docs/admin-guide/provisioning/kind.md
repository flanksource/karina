# Kind Quickstart

### 1) Install karina

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/karina/releases/latest).

=== "Linux"
```bash
wget https://github.com/flanksource/karina/releases/download/v0.15.1/karina
chmod +x karina
mv karina /usr/local/bin/karina
```
=== "MacOSX"
```bash
wget https://github.com/flanksource/karina/releases/download/v0.15.1/karina_osx
chmod +x karina_osx
mv karina_osx /usr/local/bin/karina
```


#### 2) Create a configuration file:

`test-cluster.yml`

```yaml
domain: 127.0.0.1.nip.io
name: test-cluster
calico:
  version: v3.8.2
kubernetes:
  version: v1.16.9
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

#### 3) Generate the necessary CA's:

```shell
# generate CA for kubernetes api-server authentication
karina ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1

# generate ingressCA for ingress certificates
karina ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1

```

#### 4) Provision the cluster in kind:

```shell
karina provision kind-cluster -c test-cluster.yml
```

#### 5) Deploy the bare minimum configuration:

```shell
karina deploy phases --base --dex --calico -c test-cluster.yml
```

#### 6) Deploy everything else that may be configured:

```shell
karina deploy all
```



## Troubleshooting

KIND cluster creation issues can be debugged by specifying the `--trace` argument to `karina` during creation:

```bash
karina provision kind-cluster --trace
```

