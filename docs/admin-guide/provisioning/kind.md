# Kind Quickstart

### 1) Install platform-cli

Download the latest official binary release for your platform from the [github repository](https://github.com/flanksource/platform-cli/releases/latest).

```shell
wget https://github.com/flanksource/platform-cli/releases/download/0.11.1-646-g9bbfb5c/platform-cli_osx
chmod +x platform-cli_osx
mv platform-cli_osx /usr/localbin/platform-cli
```

#### 2) Create a configuration file:

`test-cluster.yml`

```yaml
domain: 127.0.0.1.nip.io
name: test-cluster
calico:
  version: v3.8.2
versions:
  kubernetes: v1.16.4
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
platform-cli ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1

# generate ingressCA for ingress certificates
platform-cli ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1

```

#### 4) Provision the cluster in kind:

```shell
platform-cli provision kind-cluster -c test-cluster.yml
```

#### 5) Deploy the bare minimum configuration:

```shell
platform-cli deploy phases --base --dex --calico -c test-cluster.yml
```

#### 6) Deploy everything else that may be configured:

```shell
platform-cli deploy all
```



## Troubleshooting

KIND cluster creation issues can be debugged by specifying the `--trace` argument to `platform-cli` during creation:

```bash
platform-cli provision kind-cluster --trace
```

