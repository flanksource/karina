package types

type PlatformConfig struct {
	Brand         Brand          `yaml:"brand,omitempty"`
	Version       string         `yaml:"version"`
	Velero        *Velero        `yaml:"velero,omitempty"`
	CA            *CA            `yaml:"ca"`
	CanaryChecker *CanaryChecker `yaml:"canaryChecker,omitempty"`
	Calico        *Calico        `yaml:"calico,omitempty"`
	CertManager   CertManager    `yaml:"certmanager,omitempty"`
	// The endpoint for an externally hosted consul cluster
	// that is used for master discovery
	Consul         string     `yaml:"consul"`
	Dashboard      Dashboard  `yaml:"dashboard,omitempty"`
	Dex            Dex        `yaml:"dex,omitempty"`
	Datacenter     string     `yaml:"datacenter"`
	DNS            DynamicDNS `yaml:"dns,omitempty"`
	DockerRegistry string     `yaml:"dockerRegistry,omitempty"`
	// The wildcard domain that cluster will be available at
	Domain      string      `yaml:"domain"`
	EventRouter EventRouter `yaml:"eventrouter,omitempty"`
	Harbor      *Harbor     `yaml:"harbor,omitempty"`
	// A prefix to be added to VM hostnames.
	HostPrefix            string              `yaml:"hostPrefix"`
	ImportConfigs         []string            `yaml:"importConfigs,omitempty"`
	IngressCA             *CA                 `yaml:"ingressCA"`
	GitOps                []GitOps            `yaml:"gitops,omitempty"`
	Kind                  Kind                `yaml:"kind,omitempty"`
	Kiosk                 Kiosk               `yaml:"kiosk,omitempty"`
	KubeWebView           *KubeWebView        `yaml:"kubeWebView,omitempty"`
	KubeResourceReport    *KubeResourceReport `yaml:"kubeResourceReport,omitempty"`
	Kubernetes            Kubernetes          `yaml:"kubernetes"`
	Ldap                  *Ldap               `yaml:"ldap,omitempty"`
	LocalPath             *Enabled            `yaml:"localPath,omitempty"`
	Master                VM                  `yaml:"master,omitempty"`
	Monitoring            *Monitoring         `yaml:"monitoring,omitempty"`
	Name                  string              `yaml:"name"`
	NamespaceConfigurator *Enabled            `yaml:"namespaceConfigurator,omitempty"`
	NFS                   *NFS                `yaml:"nfs,omitempty"`
	Nodes                 map[string]VM       `yaml:"workers,omitempty"`
	NodeLocalDNS          NodeLocalDNS        `yaml:"nodeLocalDNS,omitempty"`
	NSX                   *NSX                `yaml:"nsx,omitempty"`
	OAuth2Proxy           *OAuth2Proxy        `yaml:"oauth2Proxy,omitempty"`
	OPA                   *OPA                `yaml:"opa,omitempty"`
	PostgresOperator      *PostgresOperator   `yaml:"postgresOperator,omitempty"`
	PodSubnet             string              `yaml:"podSubnet"`
	Policies              []string            `yaml:"policies,omitempty"`
	// A list of strategic merge patches that will be applied to all resources created
	Patches             []string             `yaml:"patches,omitempty"`
	Quack               *Enabled             `yaml:"quack,omitempty"`
	RegistryCredentials *RegistryCredentials `yaml:"registryCredentials,omitempty"`
	Resources           map[string]string    `yaml:"resources,omitempty"`
	S3                  S3                   `yaml:"s3,omitempty"`
	S3UploadCleaner     *S3UploadCleaner     `yaml:"s3uploadCleaner,omitempty"`
	SealedSecrets       *SealedSecrets       `yaml:"sealedSecrets,omitempty"`
	ServiceSubnet       string               `yaml:"serviceSubnet"`
	SMTP                SMTP                 `yaml:"smtp,omitempty"`
	Specs               []string             `yaml:"specs,omitempty"`
	TrustedCA           string               `yaml:"trustedCA,omitempty"`
	Versions            map[string]string    `yaml:"versions,omitempty"`
	PlatformOperator    *PlatformOperator    `yaml:"platformOperator,omitempty"`
	Nginx               *Nginx               `yaml:"nginx,omitempty"`
	Minio               *Enabled             `yaml:"minio,omitempty"`
	FluentdOperator     *FluentdOperator     `yaml:"fluentd,omitempty"`
	ECK                 *ECK                 `yaml:"eck,omitempty"`
	Thanos              *Thanos              `yaml:"thanos,omitempty"`
	Filebeat            []Filebeat           `yaml:"filebeat,omitempty"`
	Journalbeat         Journalbeat          `yaml:"journalbeat,omitempty"`
	Auditbeat           Auditbeat            `yaml:"auditbeat,omitempty"`
	Packetbeat          Packetbeat           `yaml:"packetbeat,omitempty"`
	Vault               *Vault               `yaml:"vault,omitempty"`
	ConfigMapReloader   ConfigMapReloader    `yaml:"configmapReloader,omitempty"`
	Elasticsearch       *Elasticsearch       `yaml:"elasticsearch,omitempty"`
	Tekton              Tekton               `yaml:"tekton,omitempty"`
	Vsphere             *Vsphere             `yaml:"vsphere,omitempty"`
	VPA                 *VPA                 `yaml:"vpa,omitempty"`
	Test                Test                 `yaml:"test,omitempty"`
	// If true, terminate operations will return an error. Used to
	// protect stateful clusters
	TerminationProtection bool   `yaml:"terminationProtection,omitempty"`
	BootstrapToken        string `yaml:"-"`
	DryRun                bool   `yaml:"-"`
	Trace                 bool   `yaml:"-"`
	JoinEndpoint          string `yaml:"-"`
	Source                string `yaml:"-"`
	ControlPlaneEndpoint  string `yaml:"-"`
	// E2E is true if end to end tests are being run
	E2E bool `yaml:"-"`
}
