# Sealed Secrets

[Sealed Secrets controller](https://github.com/bitnami-labs/sealed-secrets) is automatically deployed by `karina deploy base`.

### Configuration

```yaml
sealedSecrets:
  version: "v0.10.0"
  certificate:
    cert: .certs/sealed-secrets-crt.pem
    privateKey: .certs/sealed-secrets-key.pem
    password: foobar
```

### Certificate generation

If certificate is not provided, sealed secrets controller will automatically generate one and will store it in a secret named `sealed-secret-keys` in the `sealed-secrets` namespace.

You can override this settings and provide your own certificate for encrypting secrets.

```bash
$ karina ca generate --name sealed-secrets \
                           --cert-path .certs/sealed-secrets-crt.pem \
                           --private-key-path .certs/sealed-secrets-key.pem \
                           --password foobar \
                           --expiry 1
```
