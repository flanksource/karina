
## PlatformConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| brand |  | [Brand](#brand) |  |
| version |  | string | Yes |
| velero |  | *[Velero](#velero) |  |
| ca |  | *[CA](#ca) | Yes |
| calico |  | [Calico](#calico) |  |
| certmanager |  | [CertManager](#certmanager) |  |
| consul | The endpoint for an externally hosted consul cluster that is used for master discovery | string | Yes |
| dashboard |  | [Dashboard](#dashboard) |  |
| datacenter |  | string | Yes |
| dns |  | *[DynamicDNS](#dynamicdns) |  |
| dockerRegistry |  | string |  |
| domain | The wildcard domain that cluster will be available at | string | Yes |
| eventRouter |  | *[Enabled](#enabled) |  |
| harbor |  | *[Harbor](#harbor) |  |
| hostPrefix | A prefix to be added to VM hostnames. | string | Yes |
| importConfigs | Paths to additional config files to be merged into this config | []string |  |
| secureConfigs | Paths to [sops](https://github.com/mozilla/sops) encrypted config files to be merged into this config. Files must be encrypted with the `yaml` format | []string |  |
| importConfigs (Deprecated) | Paths to additional config files to be merged into this config.  Please use configFrom: []file instead | []string |  |
| configFrom | List of objects containing karina config | [][ImportDirective](#importdirective) |  |
| ingressCA |  | *[CA](#ca) | Yes |
| gitops |  | [][GitOps](#gitops) |  |
| kubernetes |  | [Kubernetes](#kubernetes) | Yes |
| ldap |  | *[Ldap](#ldap) |  |
| localPath |  | *[Enabled](#enabled) |  |
| master |  | [VM](#vm) |  |
| monitoring |  | *[Monitoring](#monitoring) |  |
| name |  | string | Yes |
| namespaceConfigurator |  | *[Enabled](#enabled) |  |
| nfs |  | *[NFS](#nfs) |  |
| workers |  | map[string][VM](#vm) |  |
| nodeLocalDNS |  | [NodeLocalDNS](#nodelocaldns) |  |
| nsx |  | *[NSX](#nsx) |  |
| oauth2Proxy |  | *[OAuth2Proxy](#oauth2proxy) |  |
| opa |  | *[OPA](#opa) |  |
| postgresOperator |  | *[PostgresOperator](#postgresoperator) |  |
| podSubnet |  | string | Yes |
| policies |  | []string |  |
| patches | A list of strategic merge patches that will be applied to all resources created | []string |  |
| quack |  | *[Enabled](#enabled) |  |
| registryCredentials |  | *[RegistryCredentials](#registrycredentials) |  |
| resources |  | map[string]string |  |
| s3 |  | [S3](#s3) |  |
| sealedSecrets |  | *[SealedSecrets](#sealedsecrets) |  |
| serviceSubnet |  | string | Yes |
| smtp |  | [SMTP](#smtp) |  |
| specs |  | []string |  |
| trustedCA |  | string |  |
| versions |  | map[string]string |  |
| platformOperator |  | *[Enabled](#enabled) |  |
| nginx |  | *[Nginx](#nginx) |  |
| minio |  | *[Enabled](#enabled) |  |
| fluentd |  | *[FluentdOperator](#fluentdoperator) |  |
| eck |  | *[ECK](#eck) |  |
| thanos |  | *[Thanos](#thanos) |  |
| filebeat |  | *[Filebeat](#filebeat) |  |
| vault |  | *[Vault](#vault) |  |
| configmapReloader |  | [ConfigMapReloader](#configmapreloader) |  |
| elasticsearch |  | *[Elasticsearch](#elasticsearch) |  |
| terminationProtection | If true, terminate operations will return an error. Used to protect stateful clusters | bool |  |

## AuditConfig

AuditConfig is used to specify the audit policy file. If a policy file is specified them cluster auditing is enabled. Configure additional `--audit-log-*` flags under kubernetes.apiServerExtraArgs

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| policyFile |  | string |  |

## Brand



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string |  |
| url |  | string |  |
| logo |  | string |  |

## CA



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cert |  | string |  |
| privateKey |  | string |  |
| password |  | string |  |

## Calico



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| ipip |  | calico.IPIPMode | Yes |
| vxlan |  | calico.VXLANMode | Yes |
| version |  | string |  |
| log |  | string |  |
| bgpPeers |  | []calico.BGPPeer |  |
| bgpConfig |  | calico.BGPConfiguration |  |
| ipPools |  | []calico.IPPool |  |

## CertManager



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| vault | Details of a vault server to use for signing ingress certificates | *[VaultClient](#vaultclient) |  |

## ConfigMapReloader



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| disabled |  | bool |  |

## Connection



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url |  | string | Yes |
| user |  | string |  |
| password |  | string |  |
| port |  | string |  |
| scheme |  | string |  |
| verify |  | string |  |

## Consul



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| disabled |  | bool |  |
| bucket |  | string |  |
| backupSchedule |  | string |  |
| backupImage |  | string |  |

## DB



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host |  | string | Yes |
| username |  | string | Yes |
| password |  | string | Yes |
| port |  | int | Yes |

## Dashboard



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| Enabled |  | [Enabled](#enabled) |  |
| accessRestricted |  | [LdapAccessConfig](#ldapaccessconfig) |  |

## DynamicDNS



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| nameserver |  | string |  |
| key |  | string |  |
| keyName |  | string |  |
| algorithm |  | string |  |
| zone |  | string |  |
| accessKey |  | string |  |
| secretKey |  | string |  |
| type |  | string |  |

## ECK



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string | Yes |

## Elasticsearch



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| mem |  | *[Memory](#memory) |  |
| replicas |  | int |  |
| persistence |  | *[Persistence](#persistence) |  |
| disabled |  | bool |  |

## Enabled



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool | Yes |

## Filebeat



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| disabled |  | bool |  |
| elasticsearch |  | *[Connection](#connection) |  |
| logstash |  | *[Connection](#connection) |  |

## FluentdOperator



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string | Yes |
| elasticsearch |  | [Connection](#connection) |  |
| disableDefaultConfig |  | bool | Yes |

## GitOps



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | The name of the gitops deployment, defaults to namespace name | string |  |
| disableScanning | Do not scan container image registries to fill in the registry cache, implies `--git-read-only` (default: true) | *bool |  |
| namespace | The namespace to deploy the GitOps operator into, if empty then it will be deployed cluster-wide into kube-system | string |  |
| gitUrl | The URL to git repository to clone | string | Yes |
| gitBranch | The git branch to use (default: `master`) | string |  |
| gitPath | The path with in the git repository to look for YAML in (default: `.`) | string |  |
| gitPollInterval | The frequency with which to fetch the git repository (default: `5m0s`) | string |  |
| syncInterval | The frequency with which to sync the manifests in the repository to the cluster (default: `5m0s`) | string |  |
| gitKey | The Kubernetes secret to use for cloning, if it does not exist it will be generated (default: `flux-$name-git-deploy`) | string |  |
| knownHosts | The contents of the known_hosts file to mount into Flux and helm-operator | string |  |
| sshConfig | The contents of the ~/.ssh/config file to mount into Flux and helm-operator | string |  |
| fluxVersion | The version to use for flux (default: 1.9.0 ) | string |  |
| args | a map of args to pass to flux without -- prepended. See [fluxd](https://docs.fluxcd.io/en/1.19.0/references/daemon/) for a full list | map[string]string |  |

## Grafana



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string |  |
| disabled |  | bool |  |

## Harbor



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string |  |
| chartVersion |  | string |  |
| clairVersion |  | string | Yes |
| registryVersion |  | string | Yes |
| logLevel | Logging level for various components, valid options are `info`,`warn`,`debug` (default: `warn`) | string |  |
| db |  | *[DB](#db) |  |
| url |  | string |  |
| projects |  | map[string][HarborProject](#harborproject) |  |
| settings |  | *[HarborSettings](#harborsettings) |  |
| replicas |  | int |  |
| bucket | S3 bucket for the docker registry to use | string | Yes |

## HarborProject



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string |  |
| roles |  | map[string]string |  |

## HarborSettings



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| auth_mode |  | string |  |
| email_from |  | string |  |
| email_host |  | string |  |
| email_identity |  | string |  |
| email_password |  | string |  |
| email_insecure |  | string |  |
| email_port |  | string |  |
| email_ssl |  | *bool |  |
| email_username |  | string |  |
| ldap_url |  | string |  |
| ldap_base_dn |  | string |  |
| ldap_filter |  | string |  |
| ldap_scope |  | string |  |
| ldap_search_dn |  | string |  |
| ldap_search_password |  | string |  |
| ldap_timeout |  | string |  |
| ldap_uid |  | string |  |
| ldap_verify_cert |  | *bool |  |
| ldap_group_admin_dn |  | string |  |
| ldap_group_attribute_name |  | string |  |
| ldap_group_base_dn |  | string |  |
| ldap_group_search_filter |  | string |  |
| ldap_group_search_scope |  | string |  |
| ldap_group_membership_attribute |  | string |  |
| project_creation_restriction |  | string |  |
| read_only |  | string |  |
| self_registration |  | *bool |  |
| token_expiration |  | int |  |
| oidc_name |  | string |  |
| oidc_endpoint |  | string |  |
| oidc_client_id |  | string |  |
| oidc_client_secret |  | string |  |
| oidc_scope |  | string |  |
| oidc_verify_cert |  | string |  |
| robot_token_duration |  | int |  |

## Kubernetes



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| kubeletExtraArgs | Configure additional kubelet [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/) | map[string]string |  |
| controllerExtraArgs | Configure additional kube-controller-manager [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/) | map[string]string |  |
| schedulerExtraArgs | Configure additional kube-scheduler [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-scheduler/) | map[string]string |  |
| apiServerExtraArgs | Configure additional kube-apiserver [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) | map[string]string |  |
| etcdExtraArgs | Configure additional etcd [flags](https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/configuration.md) | map[string]string |  |
| masterIP |  | string |  |
| auditing | Configure Kubernetes auditing | [AuditConfig](#auditconfig) |  |

## Ldap



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| host |  | string |  |
| port |  | string |  |
| username |  | string |  |
| password |  | string |  |
| domain |  | string |  |
| adminGroup | Members of this group will become cluster-admins | string |  |
| userDN |  | string |  |
| groupDN |  | string |  |
| groupObjectClass | GroupObjectClass is used for searching user groups in LDAP. Default is `group` for Active Directory and `groupOfNames` for Apache DS | string |  |
| groupNameAttr | GroupNameAttr is the attribute used for returning group name in OAuth tokens. Default is `name` in ActiveDirectory and `DN` in Apache DS | string |  |
| e2e |  | [LdapE2E](#ldape2e) |  |

## LdapAccessConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled |  | bool |  |
| groups |  | []string |  |
| snippet |  | string |  |

## LdapE2E



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mock | Ff true, deploy a mock LDAP server for testing | bool |  |
| username | Username to be used for OIDC integration tests | string |  |
| password | Password to be used for or OIDC integration tests | string |  |

## Memory



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| requests |  | string |  |
| limits |  | string |  |

## Monitoring



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| alert_email |  | string |  |
| version |  | string |  |
| prometheus |  | [Prometheus](#prometheus) |  |
| grafana |  | [Grafana](#grafana) |  |
| alertMmanager |  | string |  |
| kubeStateMetrics |  | string |  |
| kubeRbacProxy |  | string |  |
| nodeExporter |  | string |  |
| addonResizer |  | string |  |
| prometheus_operator |  | string |  |
| e2e |  | [MonitoringE2E](#monitoringe2e) |  |

## MonitoringE2E



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| minAlertLevel | MinAlertLevel is the minimum alert level for which E2E tests should fail. can be can be one of critical, warning, info | string |  |

## NFS



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host |  | string |  |
| path |  | string |  |

## Nginx

Configures the Nginx Ingress Controller, the controller Docker image is forked from upstream to include more LUA packages for OAuth. <br> To configure global settings not available below, override the <b>ingress-nginx/nginx-configuration</b> configmap with settings from [here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/)

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool | Yes |
| version | The version of the nginx controller to deploy (default: `0.25.1.flanksource.1`) | string | Yes |
| disableAccessLog | Disable access logs | bool |  |
| requestBodyBuffer | Size of request body buffer (default: `16M`) | string |  |
| requestBodyMax | Max size of request body (default: `32M`) | string |  |

## NodeLocalDNS



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| dnsServer |  | string |  |
| localDNS |  | string |  |
| dnsDomain |  | string |  |

## OAuth2Proxy



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool | Yes |
| cookieSecret |  | string |  |
| version |  | string |  |
| oidcGroup |  | string |  |

## OPA



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| namespaceWhitelist |  | []string |  |
| kubeMgmtVersion |  | string |  |
| version |  | string |  |
| bundleUrl |  | string |  |
| bundlePrefix |  | string |  |
| bundleServiceName |  | string |  |
| logFormat |  | string |  |
| setDecisionLogs |  | bool |  |
| policies | Policies is a path to directory containing .rego policy files | string |  |
| logLevel | Log level for opa server, one of: `debug`,`info`,`error` (default: `error`) | string |  |
| e2e |  | [OPAE2E](#opae2e) |  |

## OPAE2E



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| fixtures |  | string |  |

## Persistence



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled | Enable persistence for Prometheus | bool |  |
| storageClass | Storage class to use. If not set default one will be used | string |  |
| capacity | Capacity. Required if persistence is enabled | string |  |

## PostgresOperator



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string | Yes |
| dbVersion |  | string |  |
| backupBucket |  | string |  |
| backupSchedule |  | string |  |
| spiloImage |  | string |  |
| backupImage |  | string |  |

## Prometheus



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string |  |
| disabled |  | bool |  |
| persistence |  | [Persistence](#persistence) |  |

## RegistryCredentials



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string |  |
| namespace |  | string |  |
| aws |  | [RegistryCredentialsECR](#registrycredentialsecr) |  |
| dockerRegistry |  | [RegistryCredentialsDPR](#registrycredentialsdpr) |  |
| gcr |  | [RegistryCredentialsGCR](#registrycredentialsgcr) |  |
| azure |  | [RegistryCredentialsACR](#registrycredentialsacr) |  |

## RegistryCredentialsACR



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled |  | bool |  |
| string |  | string |  |
| clientId |  | string |  |
| password |  | string |  |

## RegistryCredentialsDPR



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled |  | bool |  |
| server |  | string |  |
| username |  | string |  |
| password |  | string |  |

## RegistryCredentialsECR



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled |  | bool |  |
| accessKey |  | string |  |
| secretKey |  | string |  |
| secretToken |  | string |  |
| account |  | string |  |
| region |  | string |  |
| assumeRole |  | string |  |

## RegistryCredentialsGCR



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled |  | bool |  |
| url |  | string |  |
| applicationCredentials |  | string |  |

## S3



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| access_key |  | string |  |
| secret_key |  | string |  |
| bucket |  | string |  |
| region |  | string |  |
| endpoint | The endpoint at which the S3-like object storage will be available from inside the cluster e.g. if minio is deployed inside the cluster, specify: `http://minio.minio.svc:9000` | string |  |
| externalEndpoint | The endpoint at which S3 is accessible outside the cluster, When deploying locally on kind specify: *minio.127.0.0.1.nip.io* | string |  |
| csiVolumes | Whether to enable the *s3* storage class that creates persistent volumes FUSE mounted to S3 buckets | bool |  |
| kmsMasterKey | Provide a KMS Master Key | string |  |
| usePathStyle | UsePathStyle http://s3host/bucket instead of http://bucket.s3host | bool | Yes |
| skipTLSVerify | Skip TLS verify when connecting to S3 | bool | Yes |
| e2e |  | [S3E2E](#s3e2e) |  |

## S3E2E



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| minio |  | bool |  |

## SMTP



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| server |  | string |  |
| username |  | string |  |
| password |  | string |  |
| port |  | int |  |
| from |  | string |  |

## SealedSecrets



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| Enabled |  | [Enabled](#enabled) |  |
| version |  | string |  |
| certificate |  | *certs.Certificate |  |

## Thanos



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool | Yes |
| version |  | string | Yes |
| mode | Must be either `client` or `obeservability`. | string |  |
| bucket | Bucket to store metrics. Must be the same across all environments | string |  |
| clientSidecars | Only for observability mode. List of client sidecars in `<hostname>:<port>`` format | []string |  |
| enableCompactor | Only for observability mode. Disable compactor singleton if there are multiple observability clusters | bool |  |
| e2e |  | [ThanosE2E](#thanose2e) |  |

## ThanosE2E



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| server |  | string |  |

## VM

VM captures the specifications of a virtual machine

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string |  |
| prefix |  | string |  |
| count | Number of VM's to provision | int | Yes |
| template |  | string | Yes |
| cluster |  | string |  |
| folder |  | string |  |
| datastore |  | string |  |
| resourcePool |  | string |  |
| cpu |  | int32 | Yes |
| memory |  | int64 | Yes |
| networks |  | []string |  |
| disk | Size in GB of the VM root volume | int | Yes |
| tags | Tags to be applied to the VM | map[string]string |  |
| commands |  | []string |  |
| konfigadm | A path to a konfigadm specification used for configuring the VM on creation. | string |  |

## Vault



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| version |  | string | Yes |
| token | A VAULT_TOKEN to use when authenticating with Vault | string |  |
| roles | A map of PKI secret roles to create/update See [pki](https://www.vaultproject.io/api-docs/secret/pki/#createupdate-role) | map[string]map[string] |  |
| policies |  | map[string]VaultPolicy |  |
| groupMappings |  | map[string][]string |  |
| config | ExtraConfig is an escape hatch that allows writing to arbritrary vault paths | map[string]map[string] |  |
| disabled |  | bool |  |
| accessKey |  | string |  |
| secretKey |  | string |  |
| kmsKeyId | The AWS KMS ARN Id to use to unseal vault | string |  |
| region |  | string |  |
| consul |  | [Consul](#consul) |  |

## VaultClient



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| address | The address of a remote Vault server to use for signinig | string | Yes |
| path | The path to the PKI Role to use for signing ingress certificates e.g. /pki/role/ingress-ca | string | Yes |
| token | A VAULT_TOKEN to use when authenticating with Vault | string | Yes |

## VaultPolicyPath



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| capabilities |  | []string |  |
| denied_parameters |  | map[string][]string |  |
| allowed_parameters |  | map[string][]string |  |

## Velero



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disabled |  | bool |  |
| version |  | string | Yes |
| schedule |  | string |  |
| bucket |  | string |  |
| volumes |  | bool | Yes |

## Versions



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubernetes |  | string |  |
| containerRuntime |  | string |  |
| dependencies |  | map[string]string |  |

## NSX



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| loadbalancer_ip_pool |  | string |  |
| tier0 |  | string |  |
| disabled |  | bool | Yes |
| Image |  | string | Yes |
| version |  | string | Yes |
| debug | If set to true, the logging level will be set to DEBUG instead of the default INFO level. | *bool |  |
| use_stderr | If set to true, log output to standard error. | *bool |  |
| use_syslog | If set to true, use syslog for logging. | *bool |  |
| log_dir | The base directory used for relative log_file paths. | string |  |
| log_file | Name of log file to send logging output to. | string |  |
| log_rotation_file_max_mb | max MB for each compressed file. Defaults to 100 MB. log_rotation_file_max_mb = 100 | *int |  |
| log_rotation_backup_count | Total number of compressed backup files to store. Defaults to 5. | *int |  |
| nsx_python_logging_path | Specify the directory where nsx-python-logging is installed | string |  |
| nsx_cli_path | Specify the directory where nsx-cli is installed | string |  |
| nsx_v3 |  | *[NsxV3](#nsxv3) |  |
| nsx_ha |  | *[NsxHA](#nsxha) |  |
| coe |  | *[NsxCOE](#nsxcoe) |  |
| nsx_k8s |  | *[NsxK8s](#nsxk8s) |  |
| nsx_node_agent |  | *[NsxNodeAgent](#nsxnodeagent) |  |

## NsxCOE



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| adaptor | Container orchestrator adaptor to plug in. | string |  |
| cluster | Specify cluster for adaptor. | string |  |
| loglevel | Log level for NCP operations Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL | string |  |
| nsxlib_loglevel | Log level for NSX API client operations Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL | string |  |
| enable_snat | Enable SNAT for all projects in this cluster | *bool |  |
| profiling | Option to enable profiling | *bool |  |
| node_type | The type of container host node Choices: HOSTVM BAREMETAL CLOUD WCP_WORKER | string |  |
| connect_retry_timeout | The time in seconds for NCP/nsx_node_agent to recover the connection to NSX manager/container orchestrator adaptor/Hyperbus before exiting. If the value is 0, NCP/nsx_node_agent wont exit automatically when the connection check fails | *int |  |

## NsxHA



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| MasterTimeout | Time duration in seconds of mastership timeout. NCP instance will remain master for this duration after elected. Note that the heartbeat period plus the update timeout must not be greater than this period. This is done to ensure that the master instance will either confirm liveness or fail before the timeout. | *int | Yes |
| HeartbeatPeriod | Time in seconds between heartbeats for elected leader. Once an NCP instance is elected master, it will periodically confirm liveness based on this value. | *int | Yes |
| UpdateTimeout | Timeout duration in seconds for update to election resource. The default value is calculated by subtracting heartbeat period from master timeout. If the update request does not complete before the timeout it will be aborted. Used for master heartbeats to ensure that the update fstructs:shes or is aborted before the master timeout occurs. | *int | Yes |

## NsxK8s



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| apiserver_host_ip | Kubernetes API server IP address. | string |  |
| apiserver_host_port | Kubernetes API server port. | string |  |
| client_token_file | Full path of the Token file to use for authenticating with the k8s API server. | string |  |
| client_cert_file | Full path of the client certificate file to use for authenticating with the k8s API server. It must be specified together with \"client_private_key_file\". | string |  |
| client_private_key_file |  | string |  |
| ca_file | Specify a CA bundle file to use in verifying the k8s API server certificate. | string |  |
| ingress_mode | Specify whether ingress controllers are expected to be deployed in hostnework mode or as regular pods externally accessed via NAT Choices: hostnetwork nat | string |  |
| loglevel | Log level for the kubernetes adaptor Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL | string |  |
| http_ingress_port |  | *int |  |
| https_ingress_port | The default HTTPS ingress port | *int |  |
| resource_watcher_thread_pool_size | Specify thread pool size to process resource events | *int |  |
| http_and_https_ingress_ip | User specified IP address for HTTP and HTTPS ingresses nolint: golint, stylecheck | string |  |
| enable_nsx_netif_crd | Set this to True to enable NCP to create segment port for VM through NsxNetworkInterface CRD. | *bool |  |
| baseline_policy_type | Option to set the type of baseline cluster policy. ALLOW_CLUSTER creates an explicit baseline policy to allow any pod to communicate any other pod within the cluster. ALLOW_NAMESPACE creates an explicit baseline policy to allow pods within the same namespace to communicate with each other. By default, no baseline rule will be created and the cluster will assume the default behavior as specified by the backend. Choices: <None> allow_cluster allow_namespace | string |  |

## NsxNodeAgent



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| log_level | The log level of NSX RPC library Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL | string |  |
| ovs_bridge | OVS bridge name | string |  |
| ovs_uplink_port | The OVS uplink OpenFlow port where to apply the NAT rules to. | string |  |
| config_retry_timeout | The time in seconds for nsx_node_agent to wait CIF config from HyperBus before returning to CNI | *int |  |
| config_reuse_backoff_time | The time in seconds for nsx_node_agent to backoff before re-using an existing cached CIF to serve CNI request. Must be less than config_retry_timeout. | *int |  |

## NsxV3



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| nsx_api_user |  | string |  |
| nsx_api_password |  | string |  |
| policy_nsxapi |  | *bool |  |
| nsx_api_cert_file | Path to NSX client certificate file. If specified, the nsx_api_user and nsx_api_password options will be ignored. Must be specified along with nsx_api_private_key_file option | string |  |
| nsx_api_private_key_file | Path to NSX client private key file. If specified, the nsx_api_user and nsx_api_password options will be ignored. Must be specified along with nsx_api_cert_file option | string |  |
| nsx_api_managers | IP address of one or more NSX managers separated by commas. The IP address should be of the form: [<scheme>://]<ip_adress>[:<port>] If scheme is not provided https is used. If port is not provided port 80 is used for http and port 443 for https. | []string |  |
| cluster_unavailable_retry | If True, skip fatal errors when no endpoint in the NSX management cluster is available to serve a request, and retry the request instead | *bool |  |
| retries | Maximum number of times to retry API requests upon stale revision errors. | *int |  |
| ca_file | Specify one or a list of CA bundle files to use in verifying the NSX Manager server certificate. This option is ignored if \"insecure\" is set to True. If \"insecure\" is set to False and ca_file is unset, the system root CAs will be used to verify the server certificate. | []string |  |
| insecure | If true, the NSX Manager server certificate is not verified. If false the CA bundle specified via \"ca_file\" will be used or if unset the default system root CAs will be used. | *bool |  |
| http_timeout | The time in seconds before aborting a HTTP connection to a NSX manager. | *int |  |
| http_read_timeout | The time in seconds before aborting a HTTP read response from a NSX manager. | *int |  |
| http_retries | Maximum number of times to retry a HTTP connection. | *int |  |
| concurrent_connections | Maximum concurrent connections to each NSX manager. | *int |  |
| conn_idlt_timeout | The amount of time in seconds to wait before ensuring connectivity to the NSX manager if no manager connection has been used. | *int |  |
| redirects | Number of times a HTTP redirect should be followed. | *int |  |
| subnet_prefix | Subnet prefix of IP block. | *int |  |
| log_dropped_traffic | Indicates whether distributed firewall DENY rules are logged. | *bool |  |
| use_native_loadbalancer | Option to use native load balancer or not | *bool |  |
| l_4_lb_auto_scaling | Option to auto scale layer 4 load balancer or not. If set to True, NCP will create additional LB when necessary upon K8s Service of type LB creation/update. | *bool |  |
| default_ingress_class_nsx | Option to use native load balancer or not when ingress class annotation is missing. Only effective if use_native_loadbalancer is set to true | *bool |  |
| lb_default_cert_path | Path to the default certificate file for HTTPS load balancing. Must be specified along with lb_priv_key_path option | string |  |
| lb_priv_key_path | Path to the private key file for default certificate for HTTPS load balancing. Must be specified along with lb_default_cert_path option | string |  |
| pool_algorithm | Option to set load balancing algorithm in load balancer pool object. Choices: ROUND_ROBIN LEAST_CONNECTION IP_HASH WEIGHTED_ROUND_ROBIN | string |  |
| service_size | Option to set load balancer service size. MEDIUM Edge VM (4 vCPU, 8GB) only supports SMALL LB. LARGE Edge VM (8 vCPU, 16GB) only supports MEDIUM and SMALL LB. Bare Metal Edge (IvyBridge, 2 socket, 128GB) supports LARGE, MEDIUM and SMALL LB Choices: SMALL MEDIUM LARGE | string |  |
| l7_persistence | Option to set load balancer persistence option. If cookie is selected, cookie persistence will be offered.If source_ip is selected, source IP persistence will be offered for ingress traffic through L7 load balancer Choices: <None> cookie source_ip | string |  |
| l7_persistence_timeout | An integer for LoadBalancer side timeout value in seconds on layer 7 persistence profile, if the profile exists. | *int |  |
| l4_persistence | Option to set load balancer persistence option. If source_ip is selected, source IP persistence will be offered for ingress traffic through L4 load balancer | string |  |
| vif_check_interval | The interval to check VIF for node. It is a workaroud for bug 2006790. Old orphan LSP may not be removed on MP, so NCP will retrieve parent VIF back once in a while. NCP will use the last created LSP from the list | *int |  |
| container_ip_blocks | Name or UUID of the container ip blocks that will be used for creating subnets. If name, it must be unique. If policy_nsxapi is enabled, it also support automatically creating the IP blocks. The definition is a comma separated list: CIDR,CIDR,... Mixing different formats (e.g. UUID,CIDR) is not supported. | []string |  |
| no_snat_ip_blocks | Name or UUID of the container ip blocks that will be used for creating subnets for no-SNAT projects. If specified, no-SNAT projects will use these ip blocks ONLY. Otherwise they will use container_ip_blocks | []string |  |
| external_ip_pools | Name or UUID of the external ip pools that will be used for allocating IP addresses which will be used for translating container IPs via SNAT rules. If policy_nsxapi is enabled, it also support automatically creating the ip pools. The definition is a comma separated list: CIDR,IP_1-IP_2,... Mixing different formats (e.g. UUID, CIDR&IP_Range) is not supported. | []string |  |
| top_tier_router | Name or UUID of the top-tier router for the container cluster network, which could be either tier0 or tier1. When policy_nsxapi is enabled, single_tier_topology is True and tier0_gateway is defined, top_tier_router value can be empty and a tier1 gateway is automatically created for the cluster | string |  |
| external_ip_pools_lb | Name or UUID of the external ip pools that will be used only for allocating IP addresses for Ingress controller and LB service | []string |  |
| overlay_tz | Name or UUID of the NSX overlay transport zone that will be used for creating logical switches for container networking. It must refer to an already existing resource on NSX and every transport node where VMs hosting containers are deployed must be enabled on this transport zone | string |  |
| x_forwarded_for | Enable X_forward_for for ingress. Available values are INSERT or REPLACE. When this config is set, if x_forwarded_for is missing, LB will add x_forwarded_for in the request header with value client ip. When x_forwarded_for is present and its set to REPLACE, LB will replace x_forwarded_for in the header to client_ip. When x_forwarded_for is present and its set to INSERT, LB will append client_ip to x_forwarded_for in the header. If not wanting to use x_forwarded_for, remove this config Choices: <None> INSERT REPLACE | string |  |
| election_profile | Name or UUID of the spoof guard switching profile that will be used by NCP for leader election | string |  |
| top_firewall_section_marker | Name or UUID of the firewall section that will be used to create firewall sections below this mark section | string |  |
| bottom_firewall_section_marker | Name or UUID of the firewall section that will be used to create firewall sections above this mark section | string |  |
| ls_replication_mode | Replication mode of container logical switch, set SOURCE for cloud as it only supports head replication mode Choices: MTEP SOURCE | string |  |
| alloc_vlan_tag | Allocate vlan ID for container interface or not. Set it to False for cloud mode. | string |  |
| search_node_tag_on | The resource which NCP will search tag 'node_name' on, to get parent VIF or transport node uuid for container LSP API context field. For HOSTVM mode, it will search tag on LSP. For BM mode, it will search tag on LSP then search TN. For CLOUD mode, it will search tag on VM. For WCP_WORKER mode, it will search TN by hostname. Choices: tag_on_lsp tag_on_tn tag_on_vm hostname_on_tn search_node_tag_on = tag_on_lsp | string |  |
| vif_app_id_type | Determines which kind of information to be used as VIF app_id. Defaults to pod_resource_key. In WCP mode, pod_uid is used. Choices: pod_resource_key pod_uid | string |  |
| snat_secondary_ips | SNAT IP to secondary IPs mapping. In the cloud case, SNAT rules are created using the PCG public or link local IPs, local IPs which will be translated to PCG secondary IPs for on-prem traffic. The secondary IPs might be used by admstructs:strator to configure on-prem firewall or other physical network services. | []string |  |
| dns_servers | If this value is not empty, NCP will append it to nameserver list | []string |  |
| enable_nsx_err_crd | Set this to True to enable NCP to report errors through NSXError CRD. | *bool |  |
| max_allowed_virtual_servers | Maximum number of virtual servers allowed to create in cluster for LoadBalancer type of services. | *int |  |
| edge_cluster | Edge cluster ID needed when creating Tier1 router for loadbalancer service. Information could be retrieved from Tier0 router | string |  |
