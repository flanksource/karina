# Templating

Config values can be templated using `env` or `template` tags.

### Env

This one will fill the environment variable value of `$AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

```yaml
vault:
  version: 1.3.2
  kmsKeyId: arn:aws:kms:us-east-1:745897381572:key/dde327f5-3b77-41b7-b42a-f9ae2270d90d
  region: us-east-1
  accessKey: !!env AWS_ACCESS_KEY_ID
  secretKey: !!env AWS_SECRET_ACCESS_KEY
```

### Template

You can use any template function defined by [gomplate](https://github.com/hairyhenderson/gomplate).

```yaml
oauth2Proxy:
  version: "v5.0.0.flanksource.1"
  oidcGroup: cn=k8s,ou=groups,dc=example,dc=com
  cookieSecret: !!template "{{ base64.Encode \"d0b0681d5babefb164b4d6e03b53967b\" }}"
```