# Configure Kubernetes SSO

### Configure LDAP in `config.yml`.

```
ldap:
  adminGroup: NA1
  username: uid=admin,ou=system
  password: secret
  port: 10636
  host: apacheds.ldap
  dn: ou=users,dc=example,dc=com
```

### Deploy ApacheDS and Dex using the following command:

```
$ platform-cli deploy stubs
$ platform-cli deploy dex
```

### Wait for it to start:

```
$ docker -n ldap get po -w
$ docker -n dex get po -w
```

### Generate kubectl config:

```
$ platform-cli kubeconfig sso
```

### Get access token:

```
$ kubelogin get-token --oidc-issuer-url=https://dex.127.0.0.1.nip.io --oidc-client-id=kubernetes --oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 --insecure-skip-tls-verify --oidc-extra-scope=email,groups,offline_access,profile,openid
```

### Set access token in ~/.kube/config

```
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server:
  name: cluster
contexts:
- context:
    cluster: cluster
    user: sso
  name: cluster
current-context: cluster
kind: Config
preferences: {}
users:
- name: cluster
  user:
    auth-provider:
      config:
        client-id: <...>
        client-secret: <...>
        extra-scopes: offline_access openid profile email groups
        idp-certificate-authority-data: <...>
        idp-issuer-url: https://dex.{domain}
        id-token: <...>
        access-token: <...>
        refresh-token: <...>
      name: oidc
```