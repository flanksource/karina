package types

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/moshloop/platform-cli/pkg/api/calico"
	yaml "gopkg.in/flanksource/yaml.v3"
)

type Enabled struct {
	Disabled bool `yaml:"disabled"`
}

type CertManager struct {
	Version string `yaml:"version"`

	// Details of a vault server to use for signing ingress certificates
	Vault *VaultClient `yaml:"vault,omitempty"`
}

type VaultClient struct {
	// The address of a remote Vault server to use for signinig
	Address string `yaml:"address"`

	// The path to the PKI Role to use for signing ingress certificates e.g. /pki/role/ingress-ca
	Path string `yaml:"path"`

	// A VAULT_TOKEN to use when authenticating with Vault
	Token string `yaml:"token"`
}

// VM captures the specifications of a virtual machine
type VM struct {
	Name   string `yaml:"name,omitempty"`
	Prefix string `yaml:"prefix,omitempty"`
	// Number of VM's to provision
	Count        int      `yaml:"count"`
	Template     string   `yaml:"template"`
	Cluster      string   `yaml:"cluster,omitempty"`
	Folder       string   `yaml:"folder,omitempty"`
	Datastore    string   `yaml:"datastore,omitempty"`
	ResourcePool string   `yaml:"resourcePool,omitempty"`
	CPUs         int32    `yaml:"cpu"`
	MemoryGB     int64    `yaml:"memory"`
	Network      []string `yaml:"networks,omitempty"`
	// Size in GB of the VM root volume
	DiskGB int `yaml:"disk"`
	// Tags to be applied to the VM
	Tags     map[string]string `yaml:"tags,omitempty"`
	Commands []string          `yaml:"commands,omitempty"`
	// A path to a konfigadm specification used for configuring the VM on creation.
	KonfigadmFile string `yaml:"konfigadm,omitempty"`
	IP            string `yaml:"-"`
}

type Calico struct {
	Disabled  bool                    `yaml:"disabled,omitempty"`
	IPIP      calico.IPIPMode         `yaml:"ipip"`
	VxLAN     calico.VXLANMode        `yaml:"vxlan"`
	Version   string                  `yaml:"version,omitempty"`
	Log       string                  `yaml:"log,omitempty"`
	BGPPeers  []calico.BGPPeer        `yaml:"bgpPeers,omitempty"`
	BGPConfig calico.BGPConfiguration `yaml:"bgpConfig,omitempty"`
	IPPools   []calico.IPPool         `yaml:"ipPools,omitempty"`
}

type OPA struct {
	Disabled           bool     `yaml:"disabled,omitempty"`
	NamespaceWhitelist []string `yaml:"namespaceWhitelist,omitempty"`
	KubeMgmtVersion    string   `yaml:"kubeMgmtVersion,omitempty"`
	Version            string   `yaml:"version,omitempty"`
	BundleURL          string   `yaml:"bundleUrl,omitempty"`
	BundlePrefix       string   `yaml:"bundlePrefix,omitempty"`
	BundleServiceName  string   `yaml:"bundleServiceName,omitempty"`
	LogFormat          string   `yaml:"logFormat,omitempty"`
	SetDecisionLogs    bool     `yaml:"setDecisionLogs,omitempty"`
	// Policies is a path to directory containing .rego policy files
	Policies string `yaml:"policies,omitempty"`
	// Log level for opa server, one of: `debug`,`info`,`error` (default: `error`)
	LogLevel string `yaml:"logLevel,omitempty"`
	E2E      OPAE2E `yaml:"e2e,omitempty"`
}

type OPAE2E struct {
	Fixtures string `yaml:"fixtures,omitempty"`
}

type Harbor struct {
	Disabled        bool   `yaml:"disabled,omitempty"`
	Version         string `yaml:"version,omitempty"`
	ChartVersion    string `yaml:"chartVersion,omitempty"`
	AdminPassword   string `yaml:"-"`
	ClairVersion    string `yaml:"clairVersion"`
	RegistryVersion string `yaml:"registryVersion"`
	// Logging level for various components, valid options are `info`,`warn`,`debug` (default: `warn`)
	LogLevel string                   `yaml:"logLevel,omitempty"`
	DB       *DB                      `yaml:"db,omitempty"`
	URL      string                   `yaml:"url,omitempty"`
	Projects map[string]HarborProject `yaml:"projects,omitempty"`
	Settings *HarborSettings          `yaml:"settings,omitempty"`
	Replicas int                      `yaml:"replicas,omitempty"`
	// S3 bucket for the docker registry to use
	Bucket string `yaml:"bucket"`
}

type HarborSettings struct {
	AuthMode                     string `json:"auth_mode,omitempty" yaml:"auth_mode,omitempty"`
	EmailFrom                    string `json:"email_from,omitempty" yaml:"email_from,omitempty"`
	EmailHost                    string `json:"email_host,omitempty" yaml:"email_host,omitempty"`
	EmailIdentity                string `json:"email_identity,omitempty" yaml:"email_identity,omitempty"`
	EmailPassword                string `json:"email_password,omitempty" yaml:"email_password,omitempty"`
	EmailInsecure                string `json:"email_insecure,omitempty" yaml:"email_insecure,omitempty"`
	EmailPort                    string `json:"email_port,omitempty" yaml:"email_port,omitempty"`
	EmailSsl                     *bool  `json:"email_ssl,omitempty" yaml:"email_ssl,omitempty"`
	EmailUsername                string `json:"email_username,omitempty" yaml:"email_username,omitempty"`
	LdapURL                      string `json:"ldap_url,omitempty" yaml:"ldap_url,omitempty"`
	LdapBaseDN                   string `json:"ldap_base_dn,omitempty" yaml:"ldap_base_dn,omitempty"`
	LdapFilter                   string `json:"ldap_filter,omitempty" yaml:"ldap_filter,omitempty"`
	LdapScope                    string `json:"ldap_scope,omitempty" yaml:"ldap_scope,omitempty"`
	LdapSearchDN                 string `json:"ldap_search_dn,omitempty" yaml:"ldap_search_dn,omitempty"`
	LdapSearchPassword           string `json:"ldap_search_password,omitempty" yaml:"ldap_search_password,omitempty"`
	LdapTimeout                  string `json:"ldap_timeout,omitempty" yaml:"ldap_timeout,omitempty"`
	LdapUID                      string `json:"ldap_uid,omitempty" yaml:"ldap_uid,omitempty"`
	LdapVerifyCert               *bool  `json:"ldap_verify_cert,omitempty" yaml:"ldap_verify_cert,omitempty"`
	LdapGroupAdminDN             string `json:"ldap_group_admin_dn,omitempty" yaml:"ldap_group_admin_dn,omitempty"`
	LdapGroupAttributeName       string `json:"ldap_group_attribute_name,omitempty" yaml:"ldap_group_attribute_name,omitempty"`
	LdapGroupBaseDN              string `json:"ldap_group_base_dn,omitempty" yaml:"ldap_group_base_dn,omitempty"`
	LdapGroupSearchFilter        string `json:"ldap_group_search_filter,omitempty" yaml:"ldap_group_search_filter,omitempty"`
	LdapGroupSearchScope         string `json:"ldap_group_search_scope,omitempty" yaml:"ldap_group_search_scope,omitempty"`
	LdapGroupMembershipAttribute string `json:"ldap_group_membership_attribute,omitempty" yaml:"ldap_group_membership_attribute,omitempty"`
	ProjectCreationRestriction   string `json:"project_creation_restriction,omitempty" yaml:"project_creation_restriction,omitempty"`
	ReadOnly                     string `json:"read_only,omitempty" yaml:"read_only,omitempty"`
	SelfRegistration             *bool  `json:"self_registration,omitempty" yaml:"self_registration,omitempty"`
	TokenExpiration              int    `json:"token_expiration,omitempty" yaml:"token_expiration,omitempty"`
	OidcName                     string `json:"oidc_name,omitempty" yaml:"oidc_name,omitempty"`
	OidcEndpoint                 string `json:"oidc_endpoint,omitempty" yaml:"oidc_endpoint,omitempty"`
	OidcClientID                 string `json:"oidc_client_id,omitempty" yaml:"oidc_client_id,omitempty"`
	OidcClientSecret             string `json:"oidc_client_secret,omitempty" yaml:"oidc_client_secret,omitempty"`
	OidcScope                    string `json:"oidc_scope,omitempty" yaml:"oidc_scope,omitempty"`
	OidcVerifyCert               string `json:"oidc_verify_cert,omitempty" yaml:"oidc_verify_cert,omitempty"`
	RobotTokenDuration           int    `json:"robot_token_duration,omitempty" yaml:"robot_token_duration,omitempty"`
}

type HarborProject struct {
	Name  string            `yaml:"name,omitempty"`
	Roles map[string]string `yaml:"roles,omitempty"`
}

type DB struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

func (db DB) GetConnectionURL(name string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", db.Username, url.QueryEscape(db.Password), db.Host, db.Port, name)
}

type PostgresOperator struct {
	Disabled       bool   `yaml:"disabled,omitempty"`
	Version        string `yaml:"version"`
	DBVersion      string `yaml:"dbVersion,omitempty"`
	BackupBucket   string `yaml:"backupBucket,omitempty"`
	BackupSchedule string `yaml:"backupSchedule,omitempty"`
	SpiloImage     string `yaml:"spiloImage,omitempty"`
	BackupImage    string `yaml:"backupImage,omitempty"`
}

type SMTP struct {
	Server   string `yaml:"server,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	From     string `yaml:"from,omitempty"`
}

type S3 struct {
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Bucket    string `yaml:"bucket,omitempty"`
	Region    string `yaml:"region,omitempty"`
	// The endpoint at which the S3-like object storage will be available from inside the cluster
	// e.g. if minio is deployed inside the cluster, specify: `http://minio.minio.svc:9000`
	Endpoint string `yaml:"endpoint,omitempty"`
	// The endpoint at which S3 is accessible outside the cluster,
	// When deploying locally on kind specify: *minio.127.0.0.1.nip.io*
	ExternalEndpoint string `yaml:"externalEndpoint,omitempty"`
	// Whether to enable the *s3* storage class that creates persistent volumes FUSE mounted to
	// S3 buckets
	CSIVolumes bool `yaml:"csiVolumes,omitempty"`
	// Provide a KMS Master Key
	KMSMasterKey string `yaml:"kmsMasterKey,omitempty"`
	// UsePathStyle http://s3host/bucket instead of http://bucket.s3host
	UsePathStyle bool `yaml:"usePathStyle"`
	// Skip TLS verify when connecting to S3
	SkipTLSVerify bool  `yaml:"skipTLSVerify"`
	E2E           S3E2E `yaml:"e2e,omitempty"`
}

type S3E2E struct {
	Minio bool `yaml:"minio,omitempty"`
}

func (s3 S3) GetExternalEndpoint() string {
	if s3.ExternalEndpoint != "" {
		return s3.ExternalEndpoint
	}
	return s3.Endpoint
}

type NFS struct {
	Host string `yaml:"host,omitempty"`
	Path string `yaml:"path,omitempty"`
}

// Configures the Nginx Ingress Controller, the controller Docker image is forked from upstream
// to include more LUA packages for OAuth. <br>
// To configure global settings not available below, override the <b>ingress-nginx/nginx-configuration</b> configmap with
// settings from [here](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/)
type Nginx struct {
	Disabled bool `yaml:"disabled"`
	// The version of the nginx controller to deploy (default: `0.25.1.flanksource.1`)
	Version string `yaml:"version"`
	// Disable access logs
	DisableAccessLog bool `yaml:"disableAccessLog,omitempty"`
	// Size of request body buffer (default: `16M`)
	RequestBodyBuffer string `yaml:"requestBodyBuffer,omitempty"`
	// Max size of request body (default: `32M`)
	RequestBodyMax string `yaml:"requestBodyMax,omitempty"`
}

type OAuth2Proxy struct {
	Disabled     bool   `yaml:"disabled"`
	CookieSecret string `yaml:"cookieSecret,omitempty"`
	Version      string `yaml:"version,omitempty"`
	OidcGroup    string `yaml:"oidcGroup,omitempty"`
}

type Ldap struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Domain   string `yaml:"domain,omitempty"`
	// Members of this group will become cluster-admins
	AdminGroup string `yaml:"adminGroup,omitempty"`
	UserDN     string `yaml:"userDN,omitempty"`
	GroupDN    string `yaml:"groupDN,omitempty"`
	// GroupObjectClass is used for searching user groups in LDAP. Default is `group` for Active Directory and `groupOfNames` for Apache DS
	GroupObjectClass string `yaml:"groupObjectClass,omitempty"`
	// GroupNameAttr is the attribute used for returning group name in OAuth tokens. Default is `name` in ActiveDirectory and `DN` in Apache DS
	GroupNameAttr string  `yaml:"groupNameAttr,omitempty"`
	E2E           LdapE2E `yaml:"e2e,omitempty"`
}

type LdapE2E struct {
	// Ff true, deploy a mock LDAP server for testing
	Mock bool `yaml:"mock,omitempty"`
	// Username to be used for OIDC integration tests
	Username string `yaml:"username,omitempty"`
	// Password to be used for or OIDC integration tests
	Password string `yaml:"password,omitempty"`
}

func (ldap Ldap) GetConnectionURL() string {
	return fmt.Sprintf("ldaps://%s:%s", ldap.Host, ldap.Port)
}

type Kubernetes struct {
	Version string `yaml:"version"`
	// Configure additional kubelet [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
	KubeletExtraArgs map[string]string `yaml:"kubeletExtraArgs,omitempty"`
	// Configure additional kube-controller-manager [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/)
	ControllerExtraArgs map[string]string `yaml:"controllerExtraArgs,omitempty"`
	// Configure additional kube-scheduler [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-scheduler/)
	SchedulerExtraArgs map[string]string `yaml:"schedulerExtraArgs,omitempty"`
	// Configure additional kube-apiserver [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/)
	APIServerExtraArgs map[string]string `yaml:"apiServerExtraArgs,omitempty"`
	// Configure additional etcd [flags](https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/configuration.md)
	EtcdExtraArgs map[string]string `yaml:"etcdExtraArgs,omitempty"`
	MasterIP      string            `yaml:"masterIP,omitempty"`
	// Configure Kubernetes auditing
	AuditConfig AuditConfig `yaml:"auditing,omitempty"`
	// EncryptionConfig is used to specify the encryption configuration file.
	EncryptionConfig EncryptionConfig `yaml:"encryption,omitempty"`
	// Configure container runtime: docker/containerd
	ContainerRuntime string `yaml:"containerRuntime"`
}

// UnmarshalYAML is used to customize the YAML unmarshalling of
// Kubernetes objects. It makes sure that if a audit policy is specified
// that a default audit-log-path will be supplied.
func (c *Kubernetes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawKubernetes Kubernetes
	raw := rawKubernetes{}

	if err := unmarshal(&raw); err != nil {
		return err
	}
	if raw.AuditConfig.PolicyFile != "" {
		if _, found := raw.APIServerExtraArgs["audit-log-path"]; !found {
			raw.APIServerExtraArgs["audit-log-path"] = "/var/log/audit/cluster-audit.log"
		}
	}

	*c = Kubernetes(raw)
	return nil
}

type Dashboard struct {
	Enabled
	AccessRestricted LdapAccessConfig `yaml:"accessRestricted,omitempty"`
}

type LdapAccessConfig struct {
	Enabled bool     `yaml:"enabled,omitempty"`
	Groups  []string `yaml:"groups,omitempty"`
	Snippet string   `yaml:"snippet,omitempty"`
}

type DynamicDNS struct {
	Disabled   bool   `yaml:"disabled,omitempty"`
	Nameserver string `yaml:"nameserver,omitempty"`
	Key        string `yaml:"key,omitempty"`
	KeyName    string `yaml:"keyName,omitempty"`
	Algorithm  string `yaml:"algorithm,omitempty"`
	Zone       string `yaml:"zone,omitempty"`
	AccessKey  string `yaml:"accessKey,omitempty"`
	SecretKey  string `yaml:"secretKey,omitempty"`
	Type       string `yaml:"type,omitempty"`
}

type Monitoring struct {
	Disabled           bool          `yaml:"disabled,omitempty"`
	AlertEmail         string        `yaml:"alert_email,omitempty"`
	Version            string        `yaml:"version,omitempty" json:"version,omitempty"`
	Prometheus         Prometheus    `yaml:"prometheus,omitempty" json:"prometheus,omitempty"`
	Grafana            Grafana       `yaml:"grafana,omitempty" json:"grafana,omitempty"`
	AlertManager       string        `yaml:"alertMmanager,omitempty"`
	KubeStateMetrics   string        `yaml:"kubeStateMetrics,omitempty"`
	KubeRbacProxy      string        `yaml:"kubeRbacProxy,omitempty"`
	NodeExporter       string        `yaml:"nodeExporter,omitempty"`
	AddonResizer       string        `yaml:"addonResizer,omitempty"`
	PrometheusOperator string        `yaml:"prometheus_operator,omitempty"`
	E2E                MonitoringE2E `yaml:"e2e,omitempty"`
}

type MonitoringE2E struct {
	// MinAlertLevel is the minimum alert level for which E2E tests should fail. can be
	// can be one of critical, warning, info
	MinAlertLevel string `yaml:"minAlertLevel,omitempty"`
}

type Prometheus struct {
	Version     string      `yaml:"version,omitempty"`
	Disabled    bool        `yaml:"disabled,omitempty"`
	Persistence Persistence `yaml:"persistence,omitempty"`
}

type Persistence struct {
	// Enable persistence for Prometheus
	Enabled bool `yaml:"enabled"`
	// Storage class to use. If not set default one will be used
	StorageClass string `yaml:"storageClass,omitempty"`
	// Capacity. Required if persistence is enabled
	Capacity string `yaml:"capacity,omitempty"`
}

type Memory struct {
	Requests string `yaml:"requests,omitempty"`
	Limits   string `yaml:"limits,omitempty"`
}

type Grafana struct {
	Version  string `yaml:"version,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
}

type Brand struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url,omitempty"`
	Logo string `yaml:"logo,omitempty"`
}

type GitOps struct {
	// The name of the gitops deployment, defaults to namespace name
	Name string `yaml:"name,omitempty"`

	// Do not scan container image registries to fill in the registry cache, implies `--git-read-only` (default: true)
	DisableScanning *bool `yaml:"disableScanning,omitempty"`

	// The namespace to deploy the GitOps operator into, if empty then it will be deployed cluster-wide into kube-system
	Namespace string `yaml:"namespace,omitempty"`

	// The URL to git repository to clone
	GitURL string `yaml:"gitUrl"`

	// The git branch to use (default: `master`)
	GitBranch string `yaml:"gitBranch,omitempty"`

	// The path with in the git repository to look for YAML in (default: `.`)
	GitPath string `yaml:"gitPath,omitempty"`

	// The frequency with which to fetch the git repository (default: `5m0s`)
	GitPollInterval string `yaml:"gitPollInterval,omitempty"`

	// The frequency with which to sync the manifests in the repository to the cluster (default: `5m0s`)
	SyncInterval string `yaml:"syncInterval,omitempty"`

	// The Kubernetes secret to use for cloning, if it does not exist it will be generated (default: `flux-$name-git-deploy`)
	GitKey string `yaml:"gitKey,omitempty"`

	// The contents of the known_hosts file to mount into Flux and helm-operator
	KnownHosts string `yaml:"knownHosts,omitempty"`

	// The contents of the ~/.ssh/config file to mount into Flux and helm-operator
	SSHConfig string `yaml:"sshConfig,omitempty"`

	// The version to use for flux (default: 1.9.0 )
	FluxVersion string `yaml:"fluxVersion,omitempty"`

	// a map of args to pass to flux without -- prepended. See [fluxd](https://docs.fluxcd.io/en/1.19.0/references/daemon/) for a full list
	Args map[string]string `yaml:"args,omitempty"`
}

type Versions struct {
	Kubernetes       string            `yaml:"kubernetes,omitempty"`
	ContainerRuntime string            `yaml:"containerRuntime,omitempty"`
	Dependencies     map[string]string `yaml:"dependencies,omitempty"`
}

type Velero struct {
	Disabled bool              `yaml:"disabled,omitempty"`
	Version  string            `yaml:"version"`
	Schedule string            `yaml:"schedule,omitempty"`
	Bucket   string            `yaml:"bucket,omitempty"`
	Volumes  bool              `yaml:"volumes"`
	Config   map[string]string `yaml:"config,omitempty"`
}

type CA struct {
	Cert       string `yaml:"cert,omitempty"`
	PrivateKey string `yaml:"privateKey,omitempty"`
	Password   string `yaml:"password,omitempty"`
}

type Thanos struct {
	Disabled bool   `yaml:"disabled"`
	Version  string `yaml:"version"`
	// Must be either `client` or `obeservability`.
	Mode string `yaml:"mode,omitempty"`
	// Bucket to store metrics. Must be the same across all environments
	Bucket string `yaml:"bucket,omitempty"`
	// Only for observability mode. List of client sidecars in `<hostname>:<port>`` format
	ClientSidecars []string `yaml:"clientSidecars,omitempty"`
	// Only for observability mode. Disable compactor singleton if there are multiple observability clusters
	EnableCompactor bool      `yaml:"enableCompactor,omitempty"`
	E2E             ThanosE2E `yaml:"e2e,omitempty"`
}

type ThanosE2E struct {
	Server string `yaml:"server,omitempty"`
}

type FluentdOperator struct {
	Disabled             bool       `yaml:"disabled,omitempty"`
	Version              string     `yaml:"version"`
	Elasticsearch        Connection `yaml:"elasticsearch,omitempty"`
	DisableDefaultConfig bool       `yaml:"disableDefaultConfig"`
}

type Filebeat struct {
	Version       string      `yaml:"version"`
	Disabled      bool        `yaml:"disabled,omitempty"`
	Name          string      `yaml:"name"`
	Index         string      `yaml:"index"`
	Prefix        string      `yaml:"prefix"`
	Elasticsearch *Connection `yaml:"elasticsearch,omitempty"`
	Logstash      *Connection `yaml:"logstash,omitempty"`
}

type Consul struct {
	Version        string `yaml:"version"`
	Disabled       bool   `yaml:"disabled,omitempty"`
	Bucket         string `yaml:"bucket,omitempty"`
	BackupSchedule string `yaml:"backupSchedule,omitempty"`
	BackupImage    string `yaml:"backupImage,omitempty"`
}

type Vault struct {
	Version string `yaml:"version"`
	// A VAULT_TOKEN to use when authenticating with Vault
	Token string `yaml:"token,omitempty"`
	// A map of PKI secret roles to create/update See [pki](https://www.vaultproject.io/api-docs/secret/pki/#createupdate-role)
	Roles         map[string]map[string]interface{} `yaml:"roles,omitempty"`
	Policies      map[string]VaultPolicy            `yaml:"policies,omitempty"`
	GroupMappings map[string][]string               `yaml:"groupMappings,omitempty"`
	// ExtraConfig is an escape hatch that allows writing to arbritrary vault paths
	ExtraConfig map[string]map[string]interface{} `yaml:"config,omitempty"`
	Disabled    bool                              `yaml:"disabled,omitempty"`
	AccessKey   string                            `yaml:"accessKey,omitempty"`
	SecretKey   string                            `yaml:"secretKey,omitempty"`
	// The AWS KMS ARN Id to use to unseal vault
	KmsKeyID string `yaml:"kmsKeyId,omitempty"`
	Region   string `yaml:"region,omitempty"`
	Consul   Consul `yaml:"consul,omitempty"`
}
type VaultPolicy map[string]VaultPolicyPath

type VaultPolicyPath struct {
	Capabilities      []string            `yaml:"capabilities,omitempty"`
	DeniedParameters  map[string][]string `yaml:"denied_parameters,omitempty"`
	AllowedParameters map[string][]string `yaml:"allowed_parameters,omitempty"`
}

func (vaultPolicy VaultPolicy) String() string {
	s := ""
	for path, policy := range vaultPolicy {
		s += fmt.Sprintf(`
		path "%s" {
			capabilities = [%s]
			denied_parameters = {
				%s
			}
			allowed_parameters {
				%s
			}
		}

		`, path, getCapabilities(policy.Capabilities),
			getParameters(policy.DeniedParameters),
			getParameters(policy.AllowedParameters))
	}
	return s
}

func getParameters(params map[string][]string) string {
	s := []string{}
	for param, keys := range params {
		s = append(s, fmt.Sprintf(`"%s" = [%s]`, param, strings.Join(wrap("\"", keys...), ",")))
	}
	return strings.Join(s, "\n")
}
func getCapabilities(capabilities []string) string {
	return strings.Join(wrap("\"", capabilities...), ",")
}

func wrap(with string, array ...string) []string {
	out := []string{}
	for _, item := range array {
		out = append(out, with+item+with)
	}
	return out
}

type ECK struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Version  string `yaml:"version"`
}

type NodeLocalDNS struct {
	Disabled  bool   `yaml:"disabled,omitempty"`
	DNSServer string `yaml:"dnsServer,omitempty"`
	LocalDNS  string `yaml:"localDNS,omitempty"`
	DNSDomain string `yaml:"dnsDomain,omitempty"`
}

type SealedSecrets struct {
	Enabled
	Version     string `yaml:"version,omitempty"`
	Certificate *CA    `yaml:"certificate,omitempty"`
}

type RegistryCredentials struct {
	Disabled              bool                   `yaml:"disabled,omitempty"`
	Version               string                 `yaml:"version,omitempty"`
	Namespace             string                 `yaml:"namespace,omitempty"`
	Aws                   RegistryCredentialsECR `yaml:"aws,omitempty"`
	DockerPrivateRegistry RegistryCredentialsDPR `yaml:"dockerRegistry,omitempty"`
	GCR                   RegistryCredentialsGCR `yaml:"gcr,omitempty"`
	ACR                   RegistryCredentialsACR `yaml:"azure,omitempty"`
}

type RegistryCredentialsECR struct {
	Enabled      bool   `yaml:"enabled,omitempty"`
	AccessKey    string `yaml:"accessKey,omitempty"`
	SecretKey    string `yaml:"secretKey,omitempty"`
	SessionToken string `yaml:"secretToken,omitempty"`
	Account      string `yaml:"account,omitempty"`
	Region       string `yaml:"region,omitempty"`
	AssumeRole   string `yaml:"assumeRole,omitempty"`
}

type RegistryCredentialsDPR struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Server   string `yaml:"server,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type RegistryCredentialsGCR struct {
	Enabled                bool   `yaml:"enabled,omitempty"`
	URL                    string `yaml:"url,omitempty"`
	ApplicationCredentials string `yaml:"applicationCredentials,omitempty"`
}

type RegistryCredentialsACR struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	URL      string `yaml:"string,omitempty"`
	ClientID string `yaml:"clientId,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type PlatformOperator struct {
	Disabled                  bool     `yaml:"disabled,omitempty"`
	Version                   string   `yaml:"version"`
	WhitelistedPodAnnotations []string `yaml:"whitelistedPodAnnotations"`
}

type Vsphere struct {
	// GOVC_USER
	Username string `yaml:"username,omitempty"`
	// GOVC_PASS
	Password string `yaml:"password,omitempty"`
	// GOVC_DATACENTER
	Datacenter string `yaml:"datacenter,omitempty"`
	// e.g. ds:///vmfs/volumes/vsan:<id>/
	DatastoreURL string `yaml:"datastoreUrl,omitempty"`
	// GOVC_DATASTORE
	Datastore string `yaml:"datastore,omitempty"`
	// GOVC_NETWORK
	Network string `yaml:"network,omitempty"`
	// Cluster for VM placement via DRS (GOVC_CLUSTER)
	Cluster string `yaml:"cluster,omitempty"`
	// GOVC_RESOURCE_POOL
	ResourcePool string `yaml:"resourcePool,omitempty"`
	//  Inventory folder (GOVC_FOLDER)
	Folder string `yaml:"folder,omitempty"`
	// GOVC_FQDN
	Hostname string `yaml:"hostname,omitempty"`
	// Version of the vSphere CSI Driver
	CSIVersion string `yaml:"csiVersion,omitempty"`
	// Version of the vSphere External Cloud Provider
	CPIVersion string `yaml:"cpiVersion,omitempty"`
	// Skip verification of server certificate
	SkipVerify bool `yaml:"verify"`
}

func (v Vsphere) GetSecret() map[string][]byte {
	return map[string][]byte{
		v.Hostname + ".username": []byte(v.Username),
		v.Hostname + ".password": []byte(v.Password),
	}
}

type Connection struct {
	URL      string `yaml:"url"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Scheme   string `yaml:"scheme,omitempty"`
	Verify   string `yaml:"verify,omitempty"`
}

// AuditConfig is used to specify the audit policy file.
// If a policy file is specified them cluster auditing is enabled.
// Configure additional `--audit-log-*` flags under kubernetes.apiServerExtraArgs
type AuditConfig struct {
	PolicyFile string `yaml:"policyFile,omitempty"`
}

// Specifies Cluster Encryption Provider Config,
// primarily by specifying the Encryption Provider Config File supplied to the cluster API Server.
type EncryptionConfig struct {
	EncryptionProviderConfigFile string `yaml:"encryptionProviderConfigFile,omitempty"`
}

type ConfigMapReloader struct {
	Version  string `yaml:"version"`
	Disabled bool   `yaml:"disabled,omitempty"`
}

type Elasticsearch struct {
	Version     string       `yaml:"version"`
	Mem         *Memory      `yaml:"mem,omitempty"`
	Replicas    int          `yaml:"replicas,omitempty"`
	Persistence *Persistence `yaml:"persistence,omitempty"`
	Disabled    bool         `yaml:"disabled,omitempty"`
}

func (c Connection) GetURL() string {
	url := c.URL
	if c.Port != "" && !strings.Contains(url, ":") {
		url = url + ":" + c.Port
	}
	if c.Scheme != "" && !strings.Contains(url, "://") {
		url = c.Scheme + "://" + url
	}
	return url
}

func (p PlatformConfig) GetImagePath(image string) string {
	if p.DockerRegistry == "" {
		return image
	}
	return p.DockerRegistry + "/" + image
}

func (p PlatformConfig) GetVMCount() int {
	count := p.Master.Count
	for _, node := range p.Nodes {
		count += node.Count
	}
	return count
}

func (p *PlatformConfig) String() string {
	data, _ := yaml.Marshal(p)
	return string(data)
}
