
## Kustomize Patches

karina provides a way to customize specification of any component deployed using a Kustomize strategic merge patches.

First create a new patch, e.g. to change the retention interval on prometheus:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: k8s
  name: k8s
  namespace: monitoring
spec:
  retention: 24h
```

Then add it to your configuration:
```yaml
patches:
  - prometheus-resources.yml
```

## Templating

Any configuration values can be templated using `env` or `template` tags  of the [flanksource/yaml](https://www.github.com/flanksource/yaml) library.

To template out environment variables `$AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

```yaml
vault:
  version: 1.3.2
  kmsKeyId: arn:aws:kms:us-east-1:745897381572:key/dde327f5-3b77-41b7-b42a-f9ae2270d90d
  region: us-east-1
  accessKey: !!env AWS_ACCESS_KEY_ID
  secretKey: !!env AWS_SECRET_ACCESS_KEY
```

You can use any template function defined by [gomplate](https://github.com/hairyhenderson/gomplate).

```yaml
oauth2Proxy:
  version: "v5.0.0.flanksource.1"
  oidcGroup: cn=k8s,ou=groups,dc=example,dc=com
  cookieSecret: !!template "{{ base64.Encode \"d0b0681d5babefb164b4d6e03b53967b\" }}"
```


## Sealed Secrets

Ensure that sealed secrets has been [deployed](sealed_secrets)

TODO

## Using a configuration hierarchy

TODO