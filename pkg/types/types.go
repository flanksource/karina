package types

import (
	"fmt"
	"net/url"

	"gopkg.in/yaml.v2"

	"github.com/moshloop/platform-cli/pkg/api/calico"
)

type Enabled struct {
	Disabled bool `yaml:"disabled"`
}

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
	IP       string            `yaml:"-"`
}

type Calico struct {
	IPIP      calico.IPIPMode         `yaml:"ipip"`
	VxLAN     calico.VXLANMode        `yaml:"vxlan"`
	Version   string                  `yaml:"version,omitempty"`
	Log       string                  `yaml:"log,omitempty"`
	BGPPeers  []calico.BGPPeer        `yaml:"bgpPeers,omitempty"`
	BGPConfig calico.BGPConfiguration `yaml:"bgpConfig,omitempty"`
	IPPools   []calico.IPPool         `yaml:"ipPools,omitempty"`
}

type OAuth2Proxy struct {
	Enabled      bool   `yaml:"enabled",omitempty`
	CookieSecret string `yaml:"cookieSecret,omitempty"`
	Version      string `yaml:"version,omitempty"`
}

type OPA struct {
	Disabled        bool   `yaml:"disabled,omitempty"`
	KubeMgmtVersion string `yaml:"kubeMgmtVersion,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

type Harbor struct {
	Disabled        bool                     `yaml:"disabled,omitempty"`
	Version         string                   `yaml:"version,omitempty"`
	ChartVersion    string                   `yaml:"chartVersion,omitempty"`
	AdminPassword   string                   `yaml:"-"`
	ClairVersion    string                   `yaml:"clairVersion"`
	RegistryVersion string                   `yaml:"registryVersion"`
	DB              *DB                      `yaml:"db,omitempty"`
	URL             string                   `yaml:"url,omitempty"`
	Projects        map[string]HarborProject `yaml:"projects,omitempty"`
	Settings        *HarborSettings          `yaml:"settings,omitempty"`
	Replicas        int                      `yaml:"replicas,omitempty"`
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

type Smtp struct {
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
	// e.g. if minio is deployed inside the cluster, specify: *http://minio.minio.svc:9000*
	Endpoint string `yaml:"endpoint,omitempty"`
	// The endpoint at which S3 is accessible outside the cluster,
	// When deploying locally on kind specify: *minio.127.0.0.1.nip.io*
	ExternalEndpoint string `yaml:"externalEndpoint,omitempty"`
	// Whether to enable the *s3* storage class that creates persistent volumes FUSE mounted to
	// S3 buckets
	CSIVolumes bool `yaml:"csiVolumes,omitempty"`
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

type Ldap struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Domain   string `yaml:"domain,omitempty"`
	// Members of this group will become cluster-admins
	AdminGroup       string `yaml:"adminGroup,omitempty"`
	UserDN           string `yaml:"userDN,omitempty"`
	GroupDN          string `yaml:"groupDN,omitempty"`
	GroupObjectClass string `yaml:"groupObjectClass,omitempty"`
	GroupNameAttr    string `yaml:"groupNameAttr,omitempty"`
}

type Kubernetes struct {
	Version          string            `yaml:"version"`
	KubeletExtraArgs map[string]string `yaml:"kubeletExtraArgs,omitempty"`
	MasterIP         string            `yaml:"masterIP,omitempty"`
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
	Disabled           bool       `yaml:"disabled,omitempty"`
	AlertEmail         string     `yaml:"alert_email,omitempty"`
	Version            string     `yaml:"version,omitempty" json:"version,omitempty"`
	Prometheus         Prometheus `yaml:"prometheus,omitempty" json:"prometheus,omitempty"`
	Grafana            Grafana    `yaml:"grafana,omitempty" json:"grafana,omitempty"`
	AlertManager       string     `yaml:"alertMmanager,omitempty"`
	KubeStateMetrics   string     `yaml:"kubeStateMetrics,omitempty"`
	KubeRbacProxy      string     `yaml:"kubeRbacProxy,omitempty"`
	NodeExporter       string     `yaml:"nodeExporter,omitempty"`
	AddonResizer       string     `yaml:"addonResizer,omitempty"`
	PrometheusOperator string     `yaml:"prometheus_operator,omitempty"`
}

type Prometheus struct {
	Version  string `yaml:"version,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
}

type Grafana struct {
	Version  string `yaml:"version,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
}

type ELK struct {
	Version      string `yaml:"version,omitempty"`
	Replicas     int    `yaml:"replicas,omitempty"`
	LogRetention string `yaml:"logRetention,omitempty"`
}

type Brand struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url,omitempty"`
	Logo string `yaml:"logo,omitempty"`
}

type GitOps struct {
	// The name of the gitops deployment, defaults to namespace name
	Name string `yaml:"name,omitempty"`

	// The namespace to deploy the GitOps operator into, if empty then it will be deployed cluster-wide into kube-system
	Namespace string `yaml:"namespace,omitempty"`

	// The URL to git repository to clone
	GitUrl string `yaml:"gitUrl"`

	// The git branch to use (default: master)
	GitBranch string `yaml:"gitBranch,omitempty"`

	// The path with in the git repository to look for YAML in (default: .)
	GitPath string `yaml:"gitPath,omitempty"`

	// The frequency with which to fetch the git repository (default: 5m0s)
	GitPollInterval string `yaml:"gitPollInterval,omitempty"`

	// The frequency with which to sync the manifests in the repository to the cluster (default: 5m0s)
	SyncInterval string `yaml:"syncInterval,omitempty"`

	// The Kubernetes secret to use for cloning, if it does not exist it will be generated (default: flux-$name-git-deploy or $GIT_SECRET_NAME)
	GitKey string `yaml:"gitKey,omitempty"`

	// The contents of the known_hosts file to mount into Flux and helm-operator
	KnownHosts string `yaml:"knownHosts,omitempty"`

	// The contents of the ~/.ssh/config file to mount into Flux and helm-operator
	SSHConfig string `yaml:"sshConfig,omitempty"`

	// The version to use for flux (default: 1.4.0 or $FLUX_VERSION)
	FluxVersion string `yaml:"fluxVersion,omitempty"`

	// a map of args to pass to flux without -- prepended
	Args map[string]string `yaml:"args,omitempty"`
}

type Versions struct {
	Kubernetes       string            `yaml:"kubernetes,omitempty"`
	ContainerRuntime string            `yaml:"containerRuntime,omitempty"`
	Dependencies     map[string]string `yaml:"dependencies,omitempty"`
}

type Velero struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Version  string `yaml:"version"`
	Schedule string `yaml:"schedule,omitempty"`
	Bucket   string `yaml:"bucket,omitempty"`
	Volumes  bool   `yaml:"volumes"`
}

type CA struct {
	Cert       string `yaml:"cert,omitempty"`
	PrivateKey string `yaml:"privateKey,omitempty"`
	Password   string `yaml:"password,omitempty"`
}

type Thanos struct {
	Disabled              bool     `yaml:"disabled"`
	Version               string   `yaml:"version"`
	Mode                  string   `yaml:"mode,omitempty"`
	ThanosSidecarEndpoint string   `yaml:"thanosSidecarEndpoint,omitempty"`
	ThanosSidecarPort     string   `yaml:"thanosSidecarPort,omitempty"`
	Bucket                string   `yaml:"bucket,omitempty"`
	ClientSidecars        []string `yaml:"clientSidecars,omitempty"`
}

type FluentdOperator struct {
	Disabled             bool       `yaml:"disabled,omitempty"`
	Version              string     `yaml:"version"`
	Elasticsearch        Connection `yaml:"elasticsearch,omitempty"`
	DisableDefaultConfig bool       `yaml:"disableDefaultConfig"`
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

type Connection struct {
	URL      string `yaml:"url"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Scheme   string `yaml:"scheme,omitempty"`
	Verify   string `yaml:"verify,omitempty"`
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
