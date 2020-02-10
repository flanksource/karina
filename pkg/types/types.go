package types

import (
	"gopkg.in/yaml.v2"

	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/api/calico"
)

type PlatformConfig struct {
	BootstrapToken        string            `yaml:"-"`
	Brand                 Brand             `yaml:"brand,omitempty"`
	Version               string            `yaml:"version,omitempty"`
	Velero                *Velero           `yaml:"velero,omitempty"`
	CA                    *CA               `yaml:"ca,omitempty"`
	Calico                Calico            `yaml:"calico,omitempty"`
	CertManager           *Enabled          `yaml:"certManager,omitempty"`
	Consul                string            `yaml:"consul,omitempty"`
	ControlPlaneEndpoint  string            `yaml:"-"`
	Dashboard             *Enabled          `yaml:"dashboard,omitempty"`
	Datacenter            string            `yaml:"datacenter,omitempty"`
	DNS                   *DynamicDNS       `yaml:"dns,omitempty"`
	DockerRegistry        string            `yaml:"dockerRegistry,omitempty"`
	Domain                string            `yaml:"domain,omitempty"`
	DryRun                bool              `yaml:"-"`
	ELK                   ELK               `yaml:"elk,omitempty"`
	EventRouter           *Enabled          `yaml:"eventRouter,omitempty"`
	Harbor                *Harbor           `yaml:"harbor,omitempty"`
	HostPrefix            string            `yaml:"hostPrefix,omitempty"`
	IngressCA             *CA               `yaml:"ingressCA,omitempty"`
	GitOps                []GitOps          `yaml:"gitops,omitempty"`
	JoinEndpoint          string            `yaml:"-"`
	Kubernetes            Kubernetes        `yaml:"kubernetes,omitempty"`
	Ldap                  *Ldap             `yaml:"ldap,omitempty"`
	LocalPath             *Enabled          `yaml:"localPath,omitempty"`
	Master                VM                `yaml:"master,omitempty"`
	Monitoring            *Monitoring       `yaml:"monitoring,omitempty"`
	Name                  string            `yaml:"name,omitempty"`
	NamespaceConfigurator *Enabled          `yaml:"namespaceConfigurator,omitempty"`
	NFS                   *NFS              `yaml:"nfs,omitempty"`
	Nodes                 map[string]VM     `yaml:"workers,omitempty"`
	NodeLocalDNS          NodeLocalDNS      `yaml:"nodeLocalDNS,omitempty"`
	NSX                   *NSX              `yaml:"nsx,omitempty"`
	OPA                   *OPA              `yaml:"opa,omitempty"`
	PGO                   *PostgresOperator `yaml:"pgo,omitempty"`
	PodSubnet             string            `yaml:"podSubnet,omitempty"`
	Policies              []string          `yaml:"policies,omitempty"`
	Quack                 *Enabled          `yaml:"quack,omitempty"`
	Resources             map[string]string `yaml:"resources,omitempty"`
	S3                    S3                `yaml:"s3,omitempty"`
	ServiceSubnet         string            `yaml:"serviceSubnet,omitempty"`
	SMTP                  Smtp              `yaml:"smtp,omitempty"`
	Source                string            `yaml:"-"`
	Specs                 []string          `yaml:"specs,omitempty"`
	TrustedCA             string            `yaml:"trustedCA,omitempty"`
	Versions              map[string]string `yaml:"versions,omitempty"`
	PlatformOperator      *Enabled          `yaml:"platformOperator,omitempty"`
	Nginx                 *Enabled          `yaml:"nginx,omitempty"`
	Minio                 *Enabled          `yaml:"minio,omitempty"`
	Thanos				  *Thanos           `yaml:"thanos,omitempty"`
	FluentdOperator       *FluentdOperator  `yaml:"fluentd-operator,omitempty"`
	ECK                   *ECK              `yaml:"eck,omitempty"`
}

type Enabled struct {
	Disabled bool `yaml:"disabled"`
}

type VM struct {
	Name         string            `yaml:"name,omitempty"`
	Prefix       string            `yaml:"prefix,omitempty"`
	Count        int               `yaml:"count,omitempty"`
	Template     string            `yaml:"template,omitempty"`
	Cluster      string            `yaml:"cluster,omitempty"`
	Folder       string            `yaml:"folder,omitempty"`
	Datastore    string            `yaml:"datastore,omitempty"`
	ResourcePool string            `yaml:"resourcePool,omitempty"`
	CPUs         int32             `yaml:"cpu,omitempty"`
	MemoryGB     int64             `yaml:"memory,omitempty"`
	Network      []string          `yaml:"networks,omitempty"`
	DiskGB       int               `yaml:"disk,omitempty"`
	Tags         map[string]string `yaml:"tags,omitempty"`
	Commands     []string          `yaml:"commands,omitempty"`
	IP           string            `yaml:"-"`
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

type OPA struct {
	Disabled        bool   `yaml:"disabled,omitempty"`
	KubeMgmtVersion string `yaml:"kubeMgmtVersion,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

type Harbor struct {
	Disabled      bool                     `yaml:"disabled,omitempty"`
	Version       string                   `yaml:"version,omitempty"`
	ChartVersion  string                   `yaml:"chartVersion,omitempty"`
	AdminPassword string                   `yaml:"-"`
	ClairVersion  string                   `yaml:"clairVersion"`
	DB            *DB                      `yaml:"db,omitempty"`
	URL           string                   `yaml:"url,omitempty"`
	Projects      map[string]HarborProject `yaml:"projects,omitempty"`
	Settings      *HarborSettings          `yaml:"settings,omitempty"`
	Replicas      int                      `yaml:"replicas,omitempty"`
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
	Host     string `yaml:"host,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     int    `yaml:"port,omitempty"`
}

type PostgresOperator struct {
	Disabled        bool                               `yaml:"disabled,omitempty"`
	Version         string                             `yaml:"version,omitempty"`
	DBVersion       string                             `yaml:"dbVersion,omitempty"`
	Password        string                             `yaml:"password,omitempty"`
	BackupBucket    string                             `yaml:"backupBucket,omitempty"`
	PrimaryStorage  string                             `yaml:"primaryStorage,omitempty"`
	XlogStorage     string                             `yaml:"xlogStorage,omitempty"`
	BackupStorage   string                             `yaml:"backupStorage,omitempty"`
	ReplicaStorage  string                             `yaml:"replicaStorage,omitempty"`
	BackrestStorage string                             `yaml:"backrestStorage,omitempty"`
	Storage         map[string]PostgresOperatorStorage `yaml:"storage,omitempty"`
}

type PostgresOperatorStorage struct {
	AccessMode   string `yaml:"AccessMode,omitempty"`
	Size         string `yaml:"Size,omitempty"`
	StorageType  string `yaml:"StorageType,omitempty"`
	StorageClass string `yaml:"StorageClass,omitempty"`
	Fsgroup      string `yaml:"Fsgroup,omitempty"`
}

type Smtp struct {
	Server   string `yaml:"server,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	From     string `yaml:"from,omitempty"`
}

type S3 struct {
	AccessKey        string `yaml:"access_key,omitempty"`
	SecretKey        string `yaml:"secret_key,omitempty"`
	Bucket           string `yaml:"bucket,omitempty"`
	Region           string `yaml:"region,omitempty"`
	Endpoint         string `yaml:"endpoint,omitempty"`
	ExternalEndpoint string `yaml:"externalEndpoint,omitempty"`
	CSIVolumes       bool   `yaml:"csiVolumes,omitempty"`
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
	Host       string `yaml:"host,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	Domain     string `yaml:"domain,omitempty"`
	AdminGroup string `yaml:"adminGroup,omitempty"`
	BindDN     string `yaml:"dn,omitempty"`
}

type Kubernetes struct {
	Version           string                          `yaml:"version,omitempty"`
	APIServer         api.KubeAPIServerConfig         `yaml:"api,omitempty"`
	Kubelet           api.KubeletConfigSpec           `yaml:"kubelet,omitempty"`
	KubeProxy         api.KubeProxyConfig             `yaml:"proxy,omitempty"`
	KubeScheduler     api.KubeSchedulerConfig         `yaml:"scheduler,omitempty"`
	ControllerManager api.KubeControllerManagerConfig `yaml:"ccm,omitempty"`
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
	Disabled         bool       `yaml:"disabled,omitempty"`
	AlertEmail       string     `yaml:"alert_email,omitempty"`
	Version          string     `yaml:"version,omitempty" json:"version,omitempty"`
	Prometheus       Prometheus `yaml:"prometheus,omitempty" json:"prometheus,omitempty"`
	Grafana          Grafana    `yaml:"grafana,omitempty" json:"grafana,omitempty"`
	AlertManager     string     `yaml:"alertMmanager,omitempty"`
	KubeStateMetrics string     `yaml:"kubeStateMetrics,omitempty"`
	KubeRbacProxy    string     `yaml:"kubeRbacProxy,omitempty"`
	NodeExporter     string     `yaml:"nodeExporter,omitempty"`
	AddonResizer     string     `yaml:"addonResizer,omitempty"`
	// Prometheus         string     `yaml:"prometheus,omitempty"`
	PrometheusOperator string `yaml:"prometheus_operator,omitempty"`
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

type Dex struct {
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

	// The URL to git repository to clone (required).
	GitUrl string `yaml:"gitUrl,omitempty"`

	// The git branch to use (default: master).
	GitBranch string `yaml:"gitBranch,omitempty"`

	// The path with in the git repository to look for YAML in (default: .).
	GitPath string `yaml:"gitPath,omitempty"`

	// The frequency with which to fetch the git repository (default: 5m0s).
	GitPollInterval string `yaml:"gitPollInterval,omitempty"`

	// The frequency with which to sync the manifests in the repository to the cluster (default: 5m0s).
	SyncInterval string `yaml:"syncInterval,omitempty"`

	// The Kubernetes secret to use for cloning, if it does not exist it will be generated (default: flux-$name-git-deploy or $GIT_SECRET_NAME).
	GitKey string `yaml:"gitKey,omitempty"`

	// The contents of the known_hosts file to mount into Flux and helm-operator.
	KnownHosts string `yaml:"knownHosts,omitempty"`

	// The contents of the ~/.ssh/config file to mount into Flux and helm-operator.
	SSHConfig string `yaml:"sshConfig,omitempty"`

	// The version to use for flux (default: 1.4.0 or $FLUX_VERSION)
	FluxVersion string `yaml:"fluxVersion,omitempty"`

	// a map of args to pass to flux without -- prepended.
	Args map[string]string `yaml:"args,omitempty"`
}

type Versions struct {
	Kubernetes       string            `yaml:"kubernetes,omitempty"`
	ContainerRuntime string            `yaml:"containerRuntime,omitempty"`
	Dependencies     map[string]string `yaml:"dependencies,omitempty"`
}

type Velero struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Version  string `yaml:"version,omitempty"`
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
	Disabled               bool   `yaml:"disabled,omitempty"`
	Version                string `yaml:"version,omitempty"`
	Mode 	               string `yaml:"mode,omitempty"`
	ThanosSidecarEndpoint  string `yaml:"thanosSidecarEndpoint,omitempty"`
	ThanosSidecarPort      string `yaml:"thanosSidecarPort,omitempty"`
	S3                    *S3     `yaml:"s3,omitempty"`
}

type FluentdOperator struct {
	Disabled  bool   `yaml:"disabled,omitempty"`
	Version   string `yaml:"version,omitempty"`
	ImageRepo string `yaml:"repository,omitempty"`
}

type ECK struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Version  string `yaml:"version,omitempty"`
}

type NodeLocalDNS struct {
	Disabled  bool   `yaml:"disabled,omitempty"`
	DNSServer string `yaml:"dnsServer,omitempty"`
	LocalDNS  string `yaml:"localDNS,omitempty"`
	DNSDomain string `yaml:"dnsDomain,omitempty"`
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
