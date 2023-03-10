package postgres

import (
	"time"

	"github.com/flanksource/kommons"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ClusterStatusUnknown etc : status of a Postgres cluster known to the operator
const (
	ClusterStatusUnknown      = ""
	ClusterStatusCreating     = "Creating"
	ClusterStatusUpdating     = "Updating"
	ClusterStatusUpdateFailed = "UpdateFailed"
	ClusterStatusSyncFailed   = "SyncFailed"
	ClusterStatusAddFailed    = "CreateFailed"
	ClusterStatusRunning      = "Running"
	ClusterStatusInvalid      = "Invalid"
)

const (
	serviceNameMaxLength   = 63
	clusterNameMaxLength   = serviceNameMaxLength - len("-repl") // nolint: unused, varcheck, deadcode
	serviceNameRegexString = `^[a-z]([-a-z0-9]*[a-z0-9])?$`      // nolint: unused, varcheck, deadcode
)

func NewPostgresql(name string) *Postgresql {
	yes := true
	return &Postgresql{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		TypeMeta: metav1.TypeMeta{
			Kind:       "postgresql",
			APIVersion: "acid.zalan.do/v1",
		},
		Spec: PostgresSpec{
			TeamID:      "postgres",
			ClusterName: name,
			DockerImage: "docker.io/flanksource/spilo:1.6-p2.flanksource",
			PostgresqlParam: PostgresqlParam{
				PgVersion:  "12",
				Parameters: make(map[string]string),
			},
			Volume: Volume{
				Size: "10Gi",
			},
			Patroni: Patroni{
				InitDB: map[string]string{
					"encoding":       "UTF8",
					"locale":         "en_US.UTF-8",
					"data-checksums": "true",
				},
				PgHba: []string{
					"hostssl all all 0.0.0.0/0 md5",
					"host    all all 0.0.0.0/0 md5",
				},
				TTL:                  30,
				LoopWait:             10,
				RetryTimeout:         10,
				MaximumLagOnFailover: 32 * 1024 * 1024,
				Slots:                make(map[string]map[string]string),
			},
			Resources: Resources{
				ResourceLimits: ResourceDescription{
					CPU:    "2",
					Memory: "2Gi",
				},
				ResourceRequests: ResourceDescription{
					CPU:    "100m",
					Memory: "100Mi",
				},
			},
			PodAnnotations:      make(map[string]string),
			ServiceAnnotations:  make(map[string]string),
			ShmVolume:           &yes,
			EnableLogicalBackup: false,
			NumberOfInstances:   2,
		},
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Postgresql defines PostgreSQL Custom Resource Definition Object.
type Postgresql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresSpec   `json:"spec"`
	Status PostgresStatus `json:"status"`
	Error  string         `json:"-"`
}

func (in Postgresql) GetObjectKind() schema.ObjectKind {
	return kommons.DynamicKind{
		APIVersion: "acid.zalan.do/v1",
		Kind:       "postgresql",
	}
}

// PostgresSpec defines the specification for the PostgreSQL TPR.
// nolint: revive
type PostgresSpec struct {
	PostgresqlParam `json:"postgresql"`
	Volume          `json:"volume,omitempty"`
	Patroni         `json:"patroni,omitempty"`
	Resources       `json:"resources,omitempty"`

	EnableConnectionPooler *bool             `json:"enableConnectionPooler,omitempty"`
	ConnectionPooler       *ConnectionPooler `json:"connectionPooler,omitempty"`

	TeamID      string `json:"teamId"`
	DockerImage string `json:"dockerImage,omitempty"`

	SpiloFSGroup *int64 `json:"spiloFSGroup,omitempty"`

	// vars that enable load balancers are pointers because it is important to know if any of them is omitted from the Postgres manifest
	// in that case the var evaluates to nil and the value is taken from the operator config
	EnableMasterLoadBalancer  *bool `json:"enableMasterLoadBalancer,omitempty"`
	EnableReplicaLoadBalancer *bool `json:"enableReplicaLoadBalancer,omitempty"`

	// deprecated load balancer settings maintained for backward compatibility
	// see "Load balancers" operator docs
	UseLoadBalancer     *bool `json:"useLoadBalancer,omitempty"`
	ReplicaLoadBalancer *bool `json:"replicaLoadBalancer,omitempty"`

	// load balancers' source ranges are the same for master and replica services
	AllowedSourceRanges []string `json:"allowedSourceRanges"`

	NumberOfInstances     int32                       `json:"numberOfInstances"`
	Users                 map[string]UserFlags        `json:"users"`
	MaintenanceWindows    []MaintenanceWindow         `json:"maintenanceWindows,omitempty"`
	Clone                 *CloneDescription           `json:"clone,omitempty"`
	ClusterName           string                      `json:"-"`
	Databases             map[string]string           `json:"databases,omitempty"`
	PreparedDatabases     map[string]PreparedDatabase `json:"preparedDatabases,omitempty"`
	Tolerations           []v1.Toleration             `json:"tolerations,omitempty"`
	Sidecars              []Sidecar                   `json:"sidecars,omitempty"`
	InitContainers        []v1.Container              `json:"initContainers,omitempty"`
	PodPriorityClassName  string                      `json:"podPriorityClassName,omitempty"`
	ShmVolume             *bool                       `json:"enableShmVolume,omitempty"`
	EnableLogicalBackup   bool                        `json:"enableLogicalBackup,omitempty"`
	LogicalBackupSchedule string                      `json:"logicalBackupSchedule,omitempty"`
	StandbyCluster        *StandbyDescription         `json:"standby,omitempty"`
	PodAnnotations        map[string]string           `json:"podAnnotations"`
	ServiceAnnotations    map[string]string           `json:"serviceAnnotations"`
	TLS                   *TLSDescription             `json:"tls"`
	Env                   []v1.EnvVar                 `json:"env,omitempty"`
	AdditionalVolumes     []AdditionalVolume          `json:"additionalVolumes,omitempty"`

	// deprecated json tags
	InitContainersOld       []v1.Container `json:"init_containers,omitempty"`
	PodPriorityClassNameOld string         `json:"pod_priority_class_name,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresqlList defines a list of PostgreSQL clusters.
type PostgresqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Postgresql `json:"items"`
}

// PreparedDatabase describes elements to be bootstrapped
type PreparedDatabase struct {
	PreparedSchemas map[string]PreparedSchema `json:"schemas,omitempty"`
	DefaultUsers    bool                      `json:"defaultUsers,omitempty" defaults:"false"`
	Extensions      map[string]string         `json:"extensions,omitempty"`
}

// PreparedSchema describes elements to be bootstrapped per schema
type PreparedSchema struct {
	DefaultRoles *bool `json:"defaultRoles,omitempty" defaults:"true"`
	DefaultUsers bool  `json:"defaultUsers,omitempty" defaults:"false"`
}

// MaintenanceWindow describes the time window when the operator is allowed to do maintenance on a cluster.
type MaintenanceWindow struct {
	Everyday  bool
	Weekday   time.Weekday
	StartTime metav1.Time // Start time
	EndTime   metav1.Time // End time
}

// Volume describes a single volume in the manifest.
type Volume struct {
	Size         string `json:"size"`
	StorageClass string `json:"storageClass"`
	SubPath      string `json:"subPath,omitempty"`
}

type AdditionalVolume struct {
	Name             string          `json:"name"`
	MountPath        string          `json:"mountPath"`
	SubPath          string          `json:"subPath"`
	TargetContainers []string        `json:"targetContainers"`
	VolumeSource     v1.VolumeSource `json:"volumeSource"`
}

// PostgresqlParam describes PostgreSQL version and pairs of configuration parameter name - values.
type PostgresqlParam struct {
	PgVersion  string            `json:"version"`
	Parameters map[string]string `json:"parameters"`
}

// ResourceDescription describes CPU and memory resources defined for a cluster.
type ResourceDescription struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// Resources describes requests and limits for the cluster resouces.
type Resources struct {
	ResourceRequests ResourceDescription `json:"requests,omitempty"`
	ResourceLimits   ResourceDescription `json:"limits,omitempty"`
}

// Patroni contains Patroni-specific configuration
type Patroni struct {
	InitDB                map[string]string            `json:"initdb"`
	PgHba                 []string                     `json:"pg_hba"`
	TTL                   uint32                       `json:"ttl"`
	LoopWait              uint32                       `json:"loop_wait"`
	RetryTimeout          uint32                       `json:"retry_timeout"`
	MaximumLagOnFailover  uint32                       `json:"maximum_lag_on_failover"` // float32 because https://github.com/kubernetes/kubernetes/issues/30213
	Slots                 map[string]map[string]string `json:"slots"`
	SynchronousMode       bool                         `json:"synchronous_mode"`
	SynchronousModeStrict bool                         `json:"synchronous_mode_strict"`
}

// StandbyCluster
type StandbyDescription struct {
	S3WalPath string `json:"s3_wal_path,omitempty"`
}

type TLSDescription struct {
	SecretName      string `json:"secretName,omitempty"`
	CertificateFile string `json:"certificateFile,omitempty"`
	PrivateKeyFile  string `json:"privateKeyFile,omitempty"`
	CAFile          string `json:"caFile,omitempty"`
	CASecretName    string `json:"caSecretName,omitempty"`
}

// CloneDescription describes which cluster the new should clone and up to which point in time
type CloneDescription struct {
	ClusterName       string `json:"cluster,omitempty"`
	UID               string `json:"uid,omitempty"`
	EndTimestamp      string `json:"timestamp,omitempty"`
	S3WalPath         string `json:"s3_wal_path,omitempty"`
	S3Endpoint        string `json:"s3_endpoint,omitempty"`
	S3AccessKeyID     string `json:"s3_access_key_id,omitempty"`
	S3SecretAccessKey string `json:"s3_secret_access_key,omitempty"`
	S3ForcePathStyle  *bool  `json:"s3_force_path_style,omitempty" defaults:"false"`
}

// Sidecar defines a container to be run in the same pod as the Postgres container.
type Sidecar struct {
	Resources   `json:"resources,omitempty"`
	Name        string             `json:"name,omitempty"`
	DockerImage string             `json:"image,omitempty"`
	Ports       []v1.ContainerPort `json:"ports,omitempty"`
	Env         []v1.EnvVar        `json:"env,omitempty"`
}

// UserFlags defines flags (such as superuser, nologin) that could be assigned to individual users
type UserFlags []string

// PostgresStatus contains status of the PostgreSQL cluster (running, creation failed etc.)
// nolint: revive
type PostgresStatus struct {
	PostgresClusterStatus string `json:"PostgresClusterStatus"`
}

// Options for connection pooler
//
// TODO: prepared snippets of configuration, one can choose via type, e.g.
// pgbouncer-large (with higher resources) or odyssey-small (with smaller
// resources)
// Type              string `json:"type,omitempty"`
//
// TODO: figure out what other important parameters of the connection pooler it
// makes sense to expose. E.g. pool size (min/max boundaries), max client
// connections etc.
type ConnectionPooler struct {
	NumberOfInstances *int32 `json:"numberOfInstances,omitempty"`
	Schema            string `json:"schema,omitempty"`
	User              string `json:"user,omitempty"`
	Mode              string `json:"mode,omitempty"`
	DockerImage       string `json:"dockerImage,omitempty"`
	MaxDBConnections  *int32 `json:"maxDBConnections,omitempty"`

	Resources `json:"resources,omitempty"`
}

// +genclient
// +genclient:onlyVerbs=get
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorConfiguration defines the specification for the OperatorConfiguration.
type OperatorConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Configuration OperatorConfigurationData `json:"configuration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorConfigurationList is used in the k8s API calls
type OperatorConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []OperatorConfiguration `json:"items"`
}

// PostgresUsersConfiguration defines the system users of Postgres.
// nolint: revive
type PostgresUsersConfiguration struct {
	SuperUsername       string `json:"super_username,omitempty"`
	ReplicationUsername string `json:"replication_username,omitempty"`
}

// KubernetesMetaConfiguration defines k8s conf required for all Postgres clusters and the operator itself
type KubernetesMetaConfiguration struct {
	PodServiceAccountName string `json:"pod_service_account_name,omitempty"`
	// TODO: change it to the proper json
	PodServiceAccountDefinition            string            `json:"pod_service_account_definition,omitempty"`
	PodServiceAccountRoleBindingDefinition string            `json:"pod_service_account_role_binding_definition,omitempty"`
	PodTerminateGracePeriod                Duration          `json:"pod_terminate_grace_period,omitempty"`
	SpiloPrivileged                        bool              `json:"spilo_privileged,omitempty"`
	SpiloFSGroup                           *int64            `json:"spilo_fsgroup,omitempty"`
	WatchedNamespace                       string            `json:"watched_namespace,omitempty"`
	EnablePodDisruptionBudget              *bool             `json:"enable_pod_disruption_budget,omitempty"`
	StorageResizeMode                      string            `json:"storage_resize_mode,omitempty"`
	EnableInitContainers                   *bool             `json:"enable_init_containers,omitempty"`
	EnableSidecars                         *bool             `json:"enable_sidecars,omitempty"`
	ClusterDomain                          string            `json:"cluster_domain,omitempty"`
	PodRoleLabel                           string            `json:"pod_role_label,omitempty"`
	ClusterLabels                          map[string]string `json:"cluster_labels,omitempty"`
	InheritedLabels                        []string          `json:"inherited_labels,omitempty"`
	DownscalerAnnotations                  []string          `json:"downscaler_annotations,omitempty"`
	ClusterNameLabel                       string            `json:"cluster_name_label,omitempty"`
	NodeReadinessLabel                     map[string]string `json:"node_readiness_label,omitempty"`
	CustomPodAnnotations                   map[string]string `json:"custom_pod_annotations,omitempty"`
	// TODO: use a proper toleration structure?
	PodToleration              map[string]string `json:"toleration,omitempty"`
	PodEnvironmentSecret       string            `json:"pod_environment_secret,omitempty"`
	PodPriorityClassName       string            `json:"pod_priority_class_name,omitempty"`
	MasterPodMoveTimeout       Duration          `json:"master_pod_move_timeout,omitempty"`
	EnablePodAntiAffinity      bool              `json:"enable_pod_antiaffinity,omitempty"`
	PodAntiAffinityTopologyKey string            `json:"pod_antiaffinity_topology_key,omitempty"`
	PodManagementPolicy        string            `json:"pod_management_policy,omitempty"`
}

// PostgresPodResourcesDefaults defines the spec of default resources
// nolint: revive
type PostgresPodResourcesDefaults struct {
	DefaultCPURequest    string `json:"default_cpu_request,omitempty"`
	DefaultMemoryRequest string `json:"default_memory_request,omitempty"`
	DefaultCPULimit      string `json:"default_cpu_limit,omitempty"`
	DefaultMemoryLimit   string `json:"default_memory_limit,omitempty"`
	MinCPULimit          string `json:"min_cpu_limit,omitempty"`
	MinMemoryLimit       string `json:"min_memory_limit,omitempty"`
}

// OperatorTimeouts defines the timeout of ResourceCheck, PodWait, ReadyWait
type OperatorTimeouts struct {
	ResourceCheckInterval  Duration `json:"resource_check_interval,omitempty"`
	ResourceCheckTimeout   Duration `json:"resource_check_timeout,omitempty"`
	PodLabelWaitTimeout    Duration `json:"pod_label_wait_timeout,omitempty"`
	PodDeletionWaitTimeout Duration `json:"pod_deletion_wait_timeout,omitempty"`
	ReadyWaitInterval      Duration `json:"ready_wait_interval,omitempty"`
	ReadyWaitTimeout       Duration `json:"ready_wait_timeout,omitempty"`
}

// LoadBalancerConfiguration defines the LB configuration
type LoadBalancerConfiguration struct {
	DBHostedZone              string            `json:"db_hosted_zone,omitempty"`
	EnableMasterLoadBalancer  bool              `json:"enable_master_load_balancer,omitempty"`
	EnableReplicaLoadBalancer bool              `json:"enable_replica_load_balancer,omitempty"`
	CustomServiceAnnotations  map[string]string `json:"custom_service_annotations,omitempty"`
}

// AWSGCPConfiguration defines the configuration for AWS
// TODO complete Google Cloud Platform (GCP) configuration
type AWSGCPConfiguration struct {
	WALES3Bucket              string `json:"wal_s3_bucket,omitempty"`
	AWSRegion                 string `json:"aws_region,omitempty"`
	WALGSBucket               string `json:"wal_gs_bucket,omitempty"`
	GCPCredentials            string `json:"gcp_credentials,omitempty"`
	LogS3Bucket               string `json:"log_s3_bucket,omitempty"`
	KubeIAMRole               string `json:"kube_iam_role,omitempty"`
	AdditionalSecretMount     string `json:"additional_secret_mount,omitempty"`
	AdditionalSecretMountPath string `json:"additional_secret_mount_path" default:"/meta/credentials"`
}

// OperatorDebugConfiguration defines options for the debug mode
type OperatorDebugConfiguration struct {
	DebugLogging   bool `json:"debug_logging,omitempty"`
	EnableDBAccess bool `json:"enable_database_access,omitempty"`
}

// TeamsAPIConfiguration defines the configuration of TeamsAPI
type TeamsAPIConfiguration struct {
	EnableTeamsAPI           bool              `json:"enable_teams_api,omitempty"`
	TeamsAPIUrl              string            `json:"teams_api_url,omitempty"`
	TeamAPIRoleConfiguration map[string]string `json:"team_api_role_configuration,omitempty"`
	EnableTeamSuperuser      bool              `json:"enable_team_superuser,omitempty"`
	EnableAdminRoleForUsers  bool              `json:"enable_admin_role_for_users,omitempty"`
	TeamAdminRole            string            `json:"team_admin_role,omitempty"`
	PamRoleName              string            `json:"pam_role_name,omitempty"`
	PamConfiguration         string            `json:"pam_configuration,omitempty"`
	ProtectedRoles           []string          `json:"protected_role_names,omitempty"`
	PostgresSuperuserTeams   []string          `json:"postgres_superuser_teams,omitempty"`
}

// LoggingRESTAPIConfiguration defines Logging API conf
type LoggingRESTAPIConfiguration struct {
	APIPort               int `json:"api_port,omitempty"`
	RingLogLines          int `json:"ring_log_lines,omitempty"`
	ClusterHistoryEntries int `json:"cluster_history_entries,omitempty"`
}

// ScalyrConfiguration defines the configuration for ScalyrAPI
type ScalyrConfiguration struct {
	ScalyrAPIKey        string `json:"scalyr_api_key,omitempty"`
	ScalyrImage         string `json:"scalyr_image,omitempty"`
	ScalyrServerURL     string `json:"scalyr_server_url,omitempty"`
	ScalyrCPURequest    string `json:"scalyr_cpu_request,omitempty"`
	ScalyrMemoryRequest string `json:"scalyr_memory_request,omitempty"`
	ScalyrCPULimit      string `json:"scalyr_cpu_limit,omitempty"`
	ScalyrMemoryLimit   string `json:"scalyr_memory_limit,omitempty"`
}

// Defines default configuration for connection pooler
type ConnectionPoolerConfiguration struct {
	NumberOfInstances    *int32 `json:"connection_pooler_number_of_instances,omitempty"`
	Schema               string `json:"connection_pooler_schema,omitempty"`
	User                 string `json:"connection_pooler_user,omitempty"`
	Image                string `json:"connection_pooler_image,omitempty"`
	Mode                 string `json:"connection_pooler_mode,omitempty"`
	MaxDBConnections     *int32 `json:"connection_pooler_max_db_connections,omitempty"`
	DefaultCPURequest    string `json:"connection_pooler_default_cpu_request,omitempty"`
	DefaultMemoryRequest string `json:"connection_pooler_default_memory_request,omitempty"`
	DefaultCPULimit      string `json:"connection_pooler_default_cpu_limit,omitempty"`
	DefaultMemoryLimit   string `json:"connection_pooler_default_memory_limit,omitempty"`
}

// OperatorLogicalBackupConfiguration defines configuration for logical backup
type OperatorLogicalBackupConfiguration struct {
	Schedule          string `json:"logical_backup_schedule,omitempty"`
	DockerImage       string `json:"logical_backup_docker_image,omitempty"`
	S3Bucket          string `json:"logical_backup_s3_bucket,omitempty"`
	S3Region          string `json:"logical_backup_s3_region,omitempty"`
	S3Endpoint        string `json:"logical_backup_s3_endpoint,omitempty"`
	S3AccessKeyID     string `json:"logical_backup_s3_access_key_id,omitempty"`
	S3SecretAccessKey string `json:"logical_backup_s3_secret_access_key,omitempty"`
	S3SSE             string `json:"logical_backup_s3_sse,omitempty"`
}

// OperatorConfigurationData defines the operation config
type OperatorConfigurationData struct {
	EnableCRDValidation     *bool    `json:"enable_crd_validation,omitempty"`
	EnableLazySpiloUpgrade  bool     `json:"enable_lazy_spilo_upgrade,omitempty"`
	KubernetesUseConfigMaps bool     `json:"kubernetes_use_configmaps,omitempty"`
	DockerImage             string   `json:"docker_image,omitempty"`
	Workers                 uint32   `json:"workers,omitempty"`
	MinInstances            int32    `json:"min_instances,omitempty"`
	MaxInstances            int32    `json:"max_instances,omitempty"`
	ResyncPeriod            Duration `json:"resync_period,omitempty"`
	RepairPeriod            Duration `json:"repair_period,omitempty"`
	SetMemoryRequestToLimit bool     `json:"set_memory_request_to_limit,omitempty"`
	ShmVolume               *bool    `json:"enable_shm_volume,omitempty"`
	// deprecated in favour of SidecarContainers
	SidecarImages              map[string]string                  `json:"sidecar_docker_images,omitempty"`
	SidecarContainers          []v1.Container                     `json:"sidecars,omitempty"`
	PostgresUsersConfiguration PostgresUsersConfiguration         `json:"users"`
	Kubernetes                 KubernetesMetaConfiguration        `json:"kubernetes"`
	PostgresPodResources       PostgresPodResourcesDefaults       `json:"postgres_pod_resources"`
	Timeouts                   OperatorTimeouts                   `json:"timeouts"`
	LoadBalancer               LoadBalancerConfiguration          `json:"load_balancer"`
	AWSGCP                     AWSGCPConfiguration                `json:"aws_or_gcp"`
	OperatorDebug              OperatorDebugConfiguration         `json:"debug"`
	TeamsAPI                   TeamsAPIConfiguration              `json:"teams_api"`
	LoggingRESTAPI             LoggingRESTAPIConfiguration        `json:"logging_rest_api"`
	Scalyr                     ScalyrConfiguration                `json:"scalyr"`
	LogicalBackup              OperatorLogicalBackupConfiguration `json:"logical_backup"`
	ConnectionPooler           ConnectionPoolerConfiguration      `json:"connection_pooler"`
}

// Duration shortens this frequently used name
type Duration time.Duration
