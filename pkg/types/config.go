// +kubebuilder:object:generate=true
package types

// +kubebuilder:skip
type PlatformConfig struct {
	Antrea            Antrea            `yaml:"antrea,omitempty" json:"antrea,omitempty"`
	ArgocdOperator    ArgocdOperator    `yaml:"argocdOperator,omitempty" json:"argocdOperator,omitempty"`
	ArgoRollouts      ArgoRollouts      `yaml:"argoRollouts,omitempty" json:"argoRollouts,omitempty"`
	Auditbeat         Auditbeat         `yaml:"auditbeat,omitempty" json:"auditbeat,omitempty"`
	Brand             Brand             `yaml:"brand,omitempty" json:"brand,omitempty"`
	CA                *CA               `yaml:"ca" json:"ca,omitempty"`
	Calico            Calico            `yaml:"calico,omitempty" json:"calico,omitempty"`
	CanaryChecker     *CanaryChecker    `yaml:"canaryChecker,omitempty" json:"canaryChecker,omitempty"`
	CertManager       CertManager       `yaml:"certmanager,omitempty" json:"certmanager,omitempty"`
	ConfigFrom        []ConfigDirective `yaml:"configFrom,omitempty" json:"configFrom,omitempty"`
	ConfigMapReloader ConfigMapReloader `yaml:"configmapReloader,omitempty" json:"configmapReloader,omitempty"`
	// The endpoint for an externally hosted consul cluster  that is used for master discovery
	Consul         string     `yaml:"consul" json:"consul,omitempty"`
	Dashboard      Dashboard  `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`
	Datacenter     string     `yaml:"datacenter" json:"datacenter,omitempty"`
	Dex            Dex        `yaml:"dex,omitempty" json:"dex,omitempty"`
	DNS            DynamicDNS `yaml:"dns,omitempty" json:"dns,omitempty"`
	DockerRegistry string     `yaml:"dockerRegistry,omitempty" json:"dockerRegistry,omitempty"`
	// The wildcard domain that cluster will be available at
	Domain        string         `yaml:"domain" json:"domain,omitempty"`
	ECK           ECK            `yaml:"eck,omitempty" json:"eck,omitempty"`
	Elasticsearch *Elasticsearch `yaml:"elasticsearch,omitempty" json:"elasticsearch,omitempty"`
	EventRouter   EventRouter    `yaml:"eventrouter,omitempty" json:"eventrouter,omitempty"`
	ExternalDNS   ExternalDNS    `yaml:"externalDns,omitempty" json:"externalDns,omitempty"`
	Filebeat      []Filebeat     `yaml:"filebeat,omitempty" json:"filebeat,omitempty"`
	Gatekeeper    Gatekeeper     `yaml:"gatekeeper,omitempty" json:"gatekeeper,omitempty"`
	GitOperator   GitOperator    `yaml:"gitOperator,omitempty" json:"gitOperator,omitempty"`
	GitOps        []GitOps       `yaml:"gitops,omitempty" json:"gitops,omitempty"`
	Harbor        *Harbor        `yaml:"harbor,omitempty" json:"harbor,omitempty"`
	// A prefix to be added to VM hostnames.
	HostPrefix string `yaml:"hostPrefix" json:"hostPrefix,omitempty"`
	// Deprecated, use configFrom instead
	ImportConfigs         []string            `yaml:"importConfigs,omitempty" json:"importConfigs,omitempty"`
	IngressCA             *CA                 `yaml:"ingressCA" json:"ingressCA,omitempty"`
	IstioOperator         IstioOperator       `yaml:"istioOperator,omitempty" json:"istioOperator,omitempty"`
	Journalbeat           Journalbeat         `yaml:"journalbeat,omitempty" json:"journalbeat,omitempty"`
	KarinaOperator        KarinaOperator      `yaml:"karinaOperator,omitempty" json:"karinaOperator,omitempty"`
	Kind                  Kind                `yaml:"kind,omitempty" json:"kind,omitempty"`
	Kiosk                 Kiosk               `yaml:"kiosk,omitempty" json:"kiosk,omitempty"`
	Kpack                 Kpack               `yaml:"kpack,omitempty" json:"kpack,omitempty"`
	KubeResourceReport    *KubeResourceReport `yaml:"kubeResourceReport,omitempty" json:"kubeResourceReport,omitempty"`
	Kubernetes            Kubernetes          `yaml:"kubernetes" json:"kubernetes,omitempty"`
	KubeWebView           *KubeWebView        `yaml:"kubeWebView,omitempty" json:"kubeWebView,omitempty"`
	Ldap                  *Ldap               `yaml:"ldap,omitempty" json:"ldap,omitempty"`
	LocalPath             *Enabled            `yaml:"localPath,omitempty" json:"localPath,omitempty"`
	LogsExporter          LogsExporter        `yaml:"logsExporter,omitempty" json:"logsExporter,omitempty"`
	Master                VM                  `yaml:"master,omitempty" json:"master,omitempty"`
	Minio                 Minio               `yaml:"minio,omitempty" json:"minio,omitempty"`
	Monitoring            *Monitoring         `yaml:"monitoring,omitempty" json:"monitoring,omitempty"`
	Name                  string              `yaml:"name" json:"name,omitempty"`
	NamespaceConfigurator *Enabled            `yaml:"namespaceConfigurator,omitempty" json:"namespaceConfigurator,omitempty"`
	NFS                   *NFS                `yaml:"nfs,omitempty" json:"nfs,omitempty"`
	Nginx                 *Nginx              `yaml:"nginx,omitempty" json:"nginx,omitempty"`
	NodeLocalDNS          NodeLocalDNS        `yaml:"nodeLocalDNS,omitempty" json:"nodeLocalDNS,omitempty"`
	Nodes                 map[string]VM       `yaml:"workers,omitempty" json:"nodes,omitempty"`
	NSX                   *NSX                `yaml:"nsx,omitempty" json:"nsx,omitempty"`
	OAuth2Proxy           *OAuth2Proxy        `yaml:"oauth2Proxy,omitempty" json:"oauth2Proxy,omitempty"`
	Packetbeat            Packetbeat          `yaml:"packetbeat,omitempty" json:"packetbeat,omitempty"`

	// A list of strategic merge patches that will be applied to all resources created, can either be a path to a file or an inline patch
	Patches             []string             `yaml:"patches,omitempty" json:"patches,omitempty"`
	PlatformOperator    PlatformOperator     `yaml:"platformOperator,omitempty" json:"platformOperator,omitempty"`
	PodSubnet           string               `yaml:"podSubnet" json:"podSubnet,omitempty"`
	Policies            []string             `yaml:"policies,omitempty" json:"policies,omitempty"`
	PostgresOperator    PostgresOperator     `yaml:"postgresOperator,omitempty" json:"postgresOperator,omitempty"`
	Quack               *Enabled             `yaml:"quack,omitempty" json:"quack,omitempty"`
	RabbitmqOperator    RabbitmqOperator     `yaml:"rabbitmqOperator,omitempty" json:"rabbitmqOperator,omitempty"`
	RedisOperator       RedisOperator        `yaml:"redisOperator,omitempty" json:"redisOperator,omitempty"`
	RegistryCredentials *RegistryCredentials `yaml:"registryCredentials,omitempty" json:"registryCredentials,omitempty"`
	Resources           map[string]string    `yaml:"resources,omitempty" json:"resources,omitempty"`
	S3                  S3                   `yaml:"s3,omitempty" json:"s3,omitempty"`
	S3UploadCleaner     *S3UploadCleaner     `yaml:"s3uploadCleaner,omitempty" json:"s3UploadCleaner,omitempty"`
	SealedSecrets       *SealedSecrets       `yaml:"sealedSecrets,omitempty" json:"sealedSecrets,omitempty"`
	ServiceSubnet       string               `yaml:"serviceSubnet" json:"serviceSubnet,omitempty"`
	SMTP                SMTP                 `yaml:"smtp,omitempty" json:"smtp,omitempty"`
	Specs               []string             `yaml:"specs,omitempty" json:"specs,omitempty"`
	Tekton              Tekton               `yaml:"tekton,omitempty" json:"tekton,omitempty"`
	TemplateOperator    TemplateOperator     `yaml:"templateOperator,omitempty" json:"templateOperator,omitempty"`

	// If true, terminate operations will return an error. Used to  protect stateful clusters
	TerminationProtection bool              `yaml:"terminationProtection,omitempty" json:"terminationProtection,omitempty"`
	Test                  Test              `yaml:"test,omitempty" json:"test,omitempty"`
	Thanos                *Thanos           `yaml:"thanos,omitempty" json:"thanos,omitempty"`
	TrustedCA             string            `yaml:"trustedCA,omitempty" json:"trustedCA,omitempty"`
	Vault                 *Vault            `yaml:"vault,omitempty" json:"vault,omitempty"`
	Velero                Velero            `yaml:"velero,omitempty" json:"velero,omitempty"`
	Version               string            `yaml:"version" json:"version,omitempty"`
	Versions              map[string]string `yaml:"versions" json:"versions,omitempty"`
	VPA                   VPA               `yaml:"vpa,omitempty" json:"vpa,omitempty"`
	Vsphere               *Vsphere          `yaml:"vsphere,omitempty" json:"vsphere,omitempty"`
	BootstrapToken        string            `yaml:"-" json:"-"`
	DryRun                bool              `yaml:"-" json:"-"`
	Trace                 bool              `yaml:"-" json:"-"`
	Prune                 bool              `yaml:"-" json:"-"`
	JoinEndpoint          string            `yaml:"-" json:"-"`
	Source                string            `yaml:"-" json:"-"`
	ControlPlaneEndpoint  string            `yaml:"-" json:"-"`
	// E2E is true if end to end tests are being run
	E2E bool `yaml:"-" json:"-"`
	// If the platform should use in cluster config
	InClusterConfig bool `yaml:"-" json:"-"`
}
