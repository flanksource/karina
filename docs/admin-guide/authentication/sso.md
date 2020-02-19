# Configure Kubernetes SSO

Configure LDAP in `config.yml`.

```
ldap:
  adminGroup: NA1
  username: uid=admin,ou=system
  password: secret
  port: 10636
  host: apacheds.ldap
  dn: ou=users,dc=example,dc=com
```

Deploy ApacheDS and Dex using the following command:

```
$ platform-cli deploy stubs
$ platform-cli deploy dex
```

Wait for it to start:

```
$ docker -n ldap get po -w
$ docker -n dex get po -w
```

Generate kubectl config:

```
$ platform-cli kubeconfig sso
```

Get access token:

```

```