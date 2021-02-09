First generate a new CA using [karina ca generate](karina_ca_generate.md)

```bash
karina ca generate --name cluster-ca \
  --cert-path cluster-ca.crt \
  --private-key-path cluster-ca.key \
  --password $CA_KEK \
  --expiry 10
```

Configure karina to use this CA when provisioning:

`karina.yml`
```yaml
ca:
  cert:  cluster-ca.crt
  privateKey: cluster-ca.key
  password: !!env CA_KEK
```

Run `karina provision` to provision a cluster and the shared CA will be injected into new instances allowing PKI based auth.



To generate a new kubeconfig file to access a cluster using a CA run:

```bash
karina kubeconfig admin --expiry 1680h --name $USER -c karina.yml
```



See [karina kubeconfig admin](karina_kubeconfig_admin.md)
