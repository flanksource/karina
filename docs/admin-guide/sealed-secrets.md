# Sealed Secrets

`karina.yml`

```yaml
sealedSecrets:
  version: "v0.10.0"
  certificate:
    cert: .certs/sealed-secrets.crt
    privateKey: .certs/sealed-secrets.key
    password: foobar
```

Deploy using:

```bash
karina deploy sealed-secrets -c karina.yml
```

### Certificate generation

If certificate is not provided, sealed secrets controller will automatically generate one and will store it in a secret named `sealed-secret-keys` in the `sealed-secrets` namespace.

You can override this settings and provide your own certificate for encrypting secrets.

```bash
karina ca generate --name sealed-secrets \
--cert-path .certs/sealed-secrets.crt \
--private-key-path .certs/sealed-secrets-key.key \
--password foobar \
--expiry 1
```

