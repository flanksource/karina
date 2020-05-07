# Vault

### Configuring a Vault Instance

1) First initialize the and seal the vault:

```yaml
vault:
  version: 1.3.3
  kmsKeyId:   # <------- Access details to the AWS KMS used for auto-unseal
  accessKey: 	#	<------- ""
  secretKey:	# <------- ""
  region: 		# <------- ""

```

```shell
platform-cli vault init
```

This will print out the Root and Recovery tokens that need to be saved

2) Then update the configuration with the Root Vault Token

```yaml
vault:
    token: !!env VAULT_TOKEN
```

3) Finally update the config to add any required policies and re-run init

```yaml
vault:
  version: 1.3.3
  token: $VAULT_TOKEN
  kmsKeyId:   # <------- Access details to the AWS KMS used for auto-unseal
  accessKey: 	#	<------- ""
  secretKey:	# <------- ""
  region: 		# <-------""
  token: $VAULT_TOKEN # <------- The root token shown in step 1, once the r
  groupMappings:
    Administrators:   # <------- AD Group Name / Role Mappings
      - admin
      - signer
  policies:
    admin:            # <------- Define roles, that are mapped to groups
      "auth/*":
        capabilities:
          - read
          - create
          - update
          - sudo
          - list
          - delete
      "sys/*":
        capabilities:
          - read
          - create
          - update
          - sudo
          - list
          - delete
    signer:
      "pki/sign/ingress":
        capabilities: ["update"]
      "pki/*":
        capabilities: ["list", "read"]
  roles:
 	 ingress: 									# <------- Configure a PKI Role for signing ingress certs
      max_ttl: 9216h #1y
      ttl: 9216h #1y
      key_type: rsa
      key_bits: 2048
      ou: Some Corp 					# <------- Default certificate request values
      organization: Some Org 	# <------- Default certificate request values
      locality: City 					# <------- Default certificate request values
      province: State 				# <------- Default certificate request values
      generate_lease: true
      require_cn: false
      allow_subdomains: true
      allowed_domains:
        - svc.cluster.local
        - wildarc.domain      # <------- The domains under which certificates can be issued

```

```shell
platform-cli vault init
```

### Configuring Cert-Manager to issue certs via Vault

```yaml
certmanager:
  vault:
    token: $VAULT_TOKEN			# <------- A token with access to the signing role
    path: pki/sign/ingress 	# <------- ingress is the name of the role specified in step 3
    address: 								# <------- https:// path to vault instance
```

Then follow the steps in [configuring automatic certificate generation](/user-guide/ingress)

### Backup / Restore

For backup/restore of vault see the underlying datastore [consul](consul)

