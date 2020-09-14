# Authentication

### Setting up SSO via LDAP

```yaml
ldap:
  adminGroup: NA1
  username: uid=admin,ou=system
  password: secret
  port: 10636
  host: apacheds.ldap
  dn: ou=users,dc=example,dc=com
```

See [login](/user-guide/login) for how to configure your kubectl client to authenticate via Dex.