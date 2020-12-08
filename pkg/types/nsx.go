package types

type NSX struct {
	LoadBalancerIPPool string `yaml:"loadbalancer_ip_pool,omitempty" json:"loadBalancerIPPool,omitempty"`
	Tier0              string `yaml:"tier0,omitempty" json:"tier0,omitempty"`
	Disabled           bool   `structs:"-" yaml:"disabled" json:"disabled,omitempty"`
	CNIDisabled        bool   `structs:"-" yaml:"cniDisabled" json:"cniDisabled,omitempty"`
	Image              string `structs:"-" yaml:"-" json:"-"`
	Version            string `structs:"-" yaml:"version" json:"version"`
	// If set to true, the logging level will be set to DEBUG instead of the
	// default INFO level.
	Debug *bool `structs:"debug,omitempty" yaml:"debug,omitempty" json:"debug,omitempty"`
	// If set to true, log output to standard error.
	UseStderr *bool `structs:"use_stderr,omitempty" yaml:"use_stderr,omitempty" json:"use_stderr,omitempty"`

	// If set to true, use syslog for logging.
	UseSyslog *bool `structs:"use_syslog,omitempty" yaml:"use_syslog,omitempty" json:"use_syslog,omitempty"`

	// The base directory used for relative log_file paths.
	LogDir string `structs:"log_dir,omitempty" yaml:"log_dir,omitempty" json:"log_dir,omitempty"`

	// Name of log file to send logging output to.
	LogFile string `structs:"log_file,omitempty" yaml:"log_file,omitempty" json:"log_file,omitempty"`

	// max MB for each compressed file. Defaults to 100 MB.
	//log_rotation_file_max_mb = 100
	LogRotationFileMaxMb *int `structs:"log_rotation_file_max_mb,omitempty" yaml:"log_rotation_file_max_mb,omitempty" json:"log_rotation_file_max_mb,omitempty"`

	// Total number of compressed backup files to store. Defaults to 5.
	LogRotationBackupCount *int `structs:"log_rotation_backup_count,omitempty" yaml:"log_rotation_backup_count,omitempty" json:"log_rotation_backup_count,omitempty"`

	// Specify the directory where nsx-python-logging is installed
	NsxPythonLoggingPath string `structs:"nsx_python_logging_path,omitempty" yaml:"nsx_python_logging_path,omitempty" json:"nsx_python_logging_path,omitempty"`

	// Specify the directory where nsx-cli is installed
	NsxCliPath string `structs:"nsx_cli_path,omitempty" yaml:"nsx_cli_path,omitempty" json:"nsx_cli_path,omitempty"`

	NsxV3 *NsxV3 `structs:"nsx_v3,omitempty" yaml:"nsx_v3,omitempty" json:"nsx_v3,omitempty"`

	NsxHA *NsxHA `structs:"ha,omitempty" yaml:"nsx_ha,omitempty" json:"nsx_ha,omitempty"`

	NsxCOE *NsxCOE `structs:"coe,omitempty" yaml:"coe,omitempty" json:"coe,omitempty"`

	NsxK8s *NsxK8s `structs:"k8s" yaml:"nsx_k8s,omitempty" json:"nsx_k8s,omitempty"`

	NsxNodeAgent *NsxNodeAgent `structs:"nsx_node_agent" yaml:"nsx_node_agent,omitempty" json:"nsx_node_agent,omitempty"`
}

type NsxV3 struct {
	NsxAPIUser   string `structs:"nsx_api_user,omitempty" yaml:"nsx_api_user,omitempty" json:"nsx_api_user,omitempty"`
	NsxAPIPass   string `structs:"nsx_api_password,omitempty" yaml:"nsx_api_password,omitempty" json:"nsx_api_password,omitempty"`
	PolicyNSXAPI *bool  `structs:"policy_nsxapi" yaml:"policy_nsxapi,omitempty" json:"policy_nsxapi,omitempty"`
	// Path to NSX client certificate file. If specified, the nsx_api_user and
	// nsx_api_password options will be ignored. Must be specified along with
	// nsx_api_private_key_file option
	NsxAPICertFile string `structs:"nsx_api_cert_file,omitempty" yaml:"nsx_api_cert_file,omitempty" json:"nsx_api_cert_file,omitempty"`

	// Path to NSX client private key file. If specified, the nsx_api_user and
	// nsx_api_password options will be ignored. Must be specified along with
	// nsx_api_cert_file option
	NsxAPIPrivateKeyFile string `structs:"nsx_api_private_key_file,omitempty" yaml:"nsx_api_private_key_file,omitempty" json:"nsx_api_private_key_file,omitempty"`

	// IP address of one or more NSX managers separated by commas. The IP
	// address should be of the form:
	// [<scheme>://]<ip_adress>[:<port>]
	// If
	// scheme is not provided https is used. If port is not provided port 80 is
	// used for http and port 443 for https.
	NsxAPIManagers []string `structs:"nsx_api_managers,omitempty" yaml:"nsx_api_managers,omitempty" json:"nsx_api_managers,omitempty"`

	// If True, skip fatal errors when no endpoint in the NSX management cluster
	// is available to serve a request, and retry the request instead
	ClusterUnavailableRetry *bool `structs:"cluster_unavailable_retry,omitempty" yaml:"cluster_unavailable_retry,omitempty" json:"cluster_unavailable_retry,omitempty"`

	// Maximum number of times to retry API requests upon stale revision errors.
	Retries *int `structs:"retries,omitempty" yaml:"retries,omitempty" json:"retries,omitempty"`

	// Specify one or a list of CA bundle files to use in verifying the NSX
	// Manager server certificate. This option is ignored if "insecure" is set
	// to True. If "insecure" is set to False and ca_file is unset, the system
	// root CAs will be used to verify the server certificate.
	CaFile []string `structs:"ca_file,omitempty" yaml:"ca_file,omitempty" json:"ca_file,omitempty"`

	// If true, the NSX Manager server certificate is not verified. If false the
	// CA bundle specified via "ca_file" will be used or if unset the default
	// system root CAs will be used.
	Insecure *bool `structs:"insecure,omitempty" yaml:"insecure,omitempty" json:"insecure,omitempty"`

	// The time in seconds before aborting a HTTP connection to a NSX manager.
	HTTPTimeout *int `structs:"http_timeout,omitempty" yaml:"http_timeout,omitempty" json:"http_timeout,omitempty"`

	// The time in seconds before aborting a HTTP read response from a NSX
	// manager.
	HTTPReadTimeout *int `structs:"http_read_timeout,omitempty" yaml:"http_read_timeout,omitempty" json:"http_read_timeout,omitempty"`

	// Maximum number of times to retry a HTTP connection.
	HTTPRetries *int `structs:"http_retries,omitempty" yaml:"http_retries,omitempty" json:"http_retries,omitempty"`

	// Maximum concurrent connections to each NSX manager.
	ConcurrentConnections *int `structs:"concurrent_connections,omitempty" yaml:"concurrent_connections,omitempty" json:"concurrent_connections,omitempty"`

	// The amount of time in seconds to wait before ensuring connectivity to the
	// NSX manager if no manager connection has been used.
	ConnIdltTimeout *int `structs:"conn_idlt_timeout,omitempty" yaml:"conn_idlt_timeout,omitempty" json:"conn_idlt_timeout,omitempty"`

	// Number of times a HTTP redirect should be followed.
	Redirects *int `structs:"redirects,omitempty" yaml:"redirects,omitempty" json:"redirects,omitempty"`

	// Subnet prefix of IP block.
	SubnetPrefix *int `structs:"subnet_prefix,omitempty" yaml:"subnet_prefix,omitempty" json:"subnet_prefix,omitempty"`

	// Indicates whether distributed firewall DENY rules are logged.
	LogDroppedTraffic *bool `structs:"log_dropped_traffic,omitempty" yaml:"log_dropped_traffic,omitempty" json:"log_dropped_traffic,omitempty"`

	// Option to use native load balancer or not
	UseNativeLoadbalancer *bool `structs:"use_native_loadbalancer,omitempty" yaml:"use_native_loadbalancer,omitempty" json:"use_native_loadbalancer,omitempty"`

	// Option to auto scale layer 4 load balancer or not. If set to True, NCP
	// will create additional LB when necessary upon K8s Service of type LB
	// creation/update.
	L4LBAutoScaling *bool `structs:"l_4_lb_auto_scaling,omitempty" yaml:"l_4_lb_auto_scaling,omitempty" json:"l_4_lb_auto_scaling,omitempty"`

	// Option to use native load balancer or not when ingress class annotation
	// is missing. Only effective if use_native_loadbalancer is set to true
	DefaultIngressClassNsx *bool `structs:"default_ingress_class_nsx,omitempty" yaml:"default_ingress_class_nsx,omitempty" json:"default_ingress_class_nsx,omitempty"`

	// Path to the default certificate file for HTTPS load balancing. Must be
	// specified along with lb_priv_key_path option
	LBDefaultCertPath string `structs:"lb_default_cert_path,omitempty" yaml:"lb_default_cert_path,omitempty" json:"lb_default_cert_path,omitempty"`

	// Path to the private key file for default certificate for HTTPS load
	// balancing. Must be specified along with lb_default_cert_path option
	LBPrivKeyPath string `structs:"lb_priv_key_path,omitempty" yaml:"lb_priv_key_path,omitempty" json:"lb_priv_key_path,omitempty"`

	// Option to set load balancing algorithm in load balancer pool object.
	// Choices: ROUND_ROBIN LEAST_CONNECTION IP_HASH WEIGHTED_ROUND_ROBIN
	PoolAlgorithm string `structs:"pool_algorithm,omitempty" yaml:"pool_algorithm,omitempty" json:"pool_algorithm,omitempty"`

	// Option to set load balancer service size. MEDIUM Edge VM (4 vCPU, 8GB)
	// only supports SMALL LB. LARGE Edge VM (8 vCPU, 16GB) only supports MEDIUM
	// and SMALL LB. Bare Metal Edge (IvyBridge, 2 socket, 128GB) supports
	// LARGE, MEDIUM and SMALL LB
	// Choices: SMALL MEDIUM LARGE
	ServiceSize string `structs:"service_size,omitempty" yaml:"service_size,omitempty" json:"service_size,omitempty"`

	// Option to set load balancer persistence option. If cookie is selected,
	// cookie persistence will be offered.If source_ip is selected, source IP
	// persistence will be offered for ingress traffic through L7 load balancer
	// Choices: <None> cookie source_ip
	L7Persistence string `structs:"l7_persistence,omitempty" yaml:"l7_persistence,omitempty" json:"l7_persistence,omitempty"`

	// An integer for LoadBalancer side timeout value in seconds on layer 7
	// persistence profile, if the profile exists.
	L7PersistenceTimeout *int `structs:"l7_persistence_timeout,omitempty" yaml:"l7_persistence_timeout,omitempty" json:"l7_persistence_timeout,omitempty"`

	// Option to set load balancer persistence option. If source_ip is selected,
	// source IP persistence will be offered for ingress traffic through L4 load
	// balancer
	L4Persistence string `structs:"l4_persistence,omitempty" yaml:"l4_persistence,omitempty" json:"l4_persistence,omitempty"`

	// The interval to check VIF for node. It is a workaroud for bug 2006790.
	// Old orphan LSP may not be removed on MP, so NCP will retrieve parent VIF
	// back once in a while. NCP will use the last created LSP from the list
	VIFCheckInterval *int `structs:"vif_check_interval,omitempty" yaml:"vif_check_interval,omitempty" json:"vif_check_interval,omitempty"`

	// Name or UUID of the container ip blocks that will be used for creating
	// subnets. If name, it must be unique. If policy_nsxapi is enabled, it also
	// support automatically creating the IP blocks. The definition is a comma
	// separated list: CIDR,CIDR,... Mixing different formats (e.g. UUID,CIDR)
	// is not supported.
	ContainerIPBlocks []string `structs:"container_ip_blocks,omitempty" yaml:"container_ip_blocks,omitempty" json:"container_ip_blocks,omitempty"`

	// Name or UUID of the container ip blocks that will be used for creating
	// subnets for no-SNAT projects. If specified, no-SNAT projects will use
	// these ip blocks ONLY. Otherwise they will use container_ip_blocks
	NoSNATIPBlocks []string `structs:"no_snat_ip_blocks,omitempty" yaml:"no_snat_ip_blocks,omitempty" json:"no_snat_ip_blocks,omitempty"`

	// Name or UUID of the external ip pools that will be used for allocating IP
	// addresses which will be used for translating container IPs via SNAT
	// rules. If policy_nsxapi is enabled, it also support automatically
	// creating the ip pools. The definition is a comma separated list:
	// CIDR,IP_1-IP_2,... Mixing different formats (e.g. UUID, CIDR&IP_Range) is
	// not supported.
	ExternalIPPools []string `structs:"external_ip_pools,omitempty" yaml:"external_ip_pools,omitempty" json:"external_ip_pools,omitempty"`

	// Name or UUID of the top-tier router for the container cluster network,
	// which could be either tier0 or tier1. When policy_nsxapi is enabled,
	// single_tier_topology is True and tier0_gateway is defined,
	// top_tier_router value can be empty and a tier1 gateway is automatically
	// created for the cluster
	TopTierRouter string `structs:"top_tier_router,omitempty" yaml:"top_tier_router,omitempty" json:"top_tier_router,omitempty"`

	// Name or UUID of the external ip pools that will be used only for
	// allocating IP addresses for Ingress controller and LB service
	ExternalIPPoolsLB []string `structs:"external_ip_pools_lb,omitempty" yaml:"external_ip_pools_lb,omitempty" json:"external_ip_pools_lb,omitempty"`

	// Name or UUID of the NSX overlay transport zone that will be used for
	// creating logical switches for container networking. It must refer to an
	// already existing resource on NSX and every transport node where VMs
	// hosting containers are deployed must be enabled on this transport zone
	OverlayTZ string `structs:"overlay_tz,omitempty" yaml:"overlay_tz,omitempty" json:"overlay_tz,omitempty"`

	// Enable X_forward_for for ingress. Available values are INSERT or REPLACE.
	// When this config is set, if x_forwarded_for is missing, LB will add
	// x_forwarded_for in the request header with value client ip. When
	// x_forwarded_for is present and its set to REPLACE, LB will replace
	// x_forwarded_for in the header to client_ip. When x_forwarded_for is
	// present and its set to INSERT, LB will append client_ip to
	// x_forwarded_for in the header. If not wanting to use x_forwarded_for,
	// remove this config
	// Choices: <None> INSERT REPLACE
	XForwardedFor string `structs:"x_forwarded_for,omitempty" yaml:"x_forwarded_for,omitempty" json:"x_forwarded_for,omitempty"`

	// Name or UUID of the spoof guard switching profile that will be used by
	// NCP for leader election
	ElectionProfile string `structs:"election_profile,omitempty" yaml:"election_profile,omitempty" json:"election_profile,omitempty"`

	// Name or UUID of the firewall section that will be used to create firewall
	// sections below this mark section
	TopFirewallSectionMarker string `structs:"top_firewall_section_marker,omitempty" yaml:"top_firewall_section_marker,omitempty" json:"top_firewall_section_marker,omitempty"`

	// Name or UUID of the firewall section that will be used to create firewall
	// sections above this mark section
	BottomFirewallSectionMarker string `structs:"bottom_firewall_section_marker,omitempty" yaml:"bottom_firewall_section_marker,omitempty" json:"bottom_firewall_section_marker,omitempty"`

	// Replication mode of container logical switch, set SOURCE for cloud as it
	// only supports head replication mode
	// Choices: MTEP SOURCE
	LSReplicationMode string `structs:"ls_replication_mode,omitempty" yaml:"ls_replication_mode,omitempty" json:"ls_replication_mode,omitempty"`

	// Allocate vlan ID for container interface or not. Set it to False for
	// cloud mode.
	AllocVlanTag string `structs:"alloc_vlan_tag,omitempty" yaml:"alloc_vlan_tag,omitempty" json:"alloc_vlan_tag,omitempty"`

	// The resource which NCP will search tag 'node_name' on, to get parent VIF
	// or transport node uuid for container LSP API context field. For HOSTVM
	// mode, it will search tag on LSP. For BM mode, it will search tag on LSP
	// then search TN. For CLOUD mode, it will search tag on VM. For WCP_WORKER
	// mode, it will search TN by hostname.
	// Choices: tag_on_lsp tag_on_tn tag_on_vm hostname_on_tn
	//search_node_tag_on = tag_on_lsp
	SearchNodeTagOn string `structs:"search_node_tag_on,omitempty" yaml:"search_node_tag_on,omitempty" json:"search_node_tag_on,omitempty"`

	// Determines which kind of information to be used as VIF app_id. Defaults
	// to pod_resource_key. In WCP mode, pod_uid is used.
	// Choices: pod_resource_key pod_uid
	VifAppIDType string `structs:"vif_app_id_type,omitempty" yaml:"vif_app_id_type,omitempty" json:"vif_app_id_type,omitempty"`

	// SNAT IP to secondary IPs mapping. In the cloud case, SNAT rules are
	// created using the PCG public or link local IPs, local IPs which will be
	// translated to PCG secondary IPs for on-prem traffic. The secondary IPs
	// might be used by admstructs:strator to configure on-prem firewall or other
	// physical network services.
	SnatSecondaryIps []string `structs:"snat_secondary_ips,omitempty" yaml:"snat_secondary_ips,omitempty" json:"snat_secondary_ips,omitempty"`

	// If this value is not empty, NCP will append it to nameserver list
	DNSServers []string `structs:"dns_servers,omitempty" yaml:"dns_servers,omitempty" json:"dns_servers,omitempty"`

	// Set this to True to enable NCP to report errors through NSXError CRD.
	EnableNsxErrCrd *bool `structs:"enable_nsx_err_crd,omitempty" yaml:"enable_nsx_err_crd,omitempty" json:"enable_nsx_err_crd,omitempty"`

	// Maximum number of virtual servers allowed to create in cluster for
	// LoadBalancer type of services.
	MaxAllowedVirtualServers *int `structs:"max_allowed_virtual_servers,omitempty" yaml:"max_allowed_virtual_servers,omitempty" json:"max_allowed_virtual_servers,omitempty"`

	// Edge cluster ID needed when creating Tier1 router for loadbalancer
	// service. Information could be retrieved from Tier0 router
	EdgeCluster string `structs:"edge_cluster,omitempty" yaml:"edge_cluster,omitempty" json:"edge_cluster,omitempty"`
}

type NsxHA struct {

	// Time duration in seconds of mastership timeout. NCP instance will remain
	// master for this duration after elected. Note that the heartbeat period
	// plus the update timeout must not be greater than this period. This is
	// done to ensure that the master instance will either confirm liveness or
	// fail before the timeout.
	MasterTimeout *int `structs:"master_timeout,omitempty" json:"master_timeout,omitempty"`

	// Time in seconds between heartbeats for elected leader. Once an NCP
	// instance is elected master, it will periodically confirm liveness based
	// on this value.
	HeartbeatPeriod *int `structs:"heartbeat_period,omitempty" json:"heartbeat_period,omitempty"`

	// Timeout duration in seconds for update to election resource. The default
	// value is calculated by subtracting heartbeat period from master timeout.
	// If the update request does not complete before the timeout it will be
	// aborted. Used for master heartbeats to ensure that the update fstructs:shes or
	// is aborted before the master timeout occurs.
	UpdateTimeout *int `structs:"update_timeout,omitempty" json:"update_timeout,omitempty"`
}

type NsxCOE struct {

	// Container orchestrator adaptor to plug in.
	Adaptor string `structs:"adaptor,omitempty" yaml:"adaptor,omitempty" json:"adaptor,omitempty"`

	// Specify cluster for adaptor.
	Cluster string `structs:"cluster,omitempty" yaml:"cluster,omitempty" json:"cluster,omitempty"`

	// Log level for NCP operations
	// Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL
	Loglevel string `structs:"loglevel,omitempty" yaml:"loglevel,omitempty" json:"loglevel,omitempty"`

	// Log level for NSX API client operations
	// Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL
	NsxlibLoglevel string `structs:"nsxlib_loglevel,omitempty" yaml:"nsxlib_loglevel,omitempty" json:"nsxlib_loglevel,omitempty"`

	// Enable SNAT for all projects in this cluster
	EnableSnat *bool `structs:"enable_snat,omitempty" yaml:"enable_snat,omitempty" json:"enable_snat,omitempty"`

	// Option to enable profiling
	Profiling *bool `structs:"profiling,omitempty" yaml:"profiling,omitempty" json:"profiling,omitempty"`

	// The type of container host node
	// Choices: HOSTVM BAREMETAL CLOUD WCP_WORKER
	NodeType string `structs:"node_type,omitempty" yaml:"node_type,omitempty" json:"node_type,omitempty"`

	// The time in seconds for NCP/nsx_node_agent to recover the connection to
	// NSX manager/container orchestrator adaptor/Hyperbus before exiting. If
	// the value is 0, NCP/nsx_node_agent wont exit automatically when the
	// connection check fails
	ConnectRetryTimeout *int `structs:"connect_retry_timeout,omitempty" yaml:"connect_retry_timeout,omitempty" json:"connect_retry_timeout,omitempty"`
}

type NsxK8s struct {
	// Kubernetes API server IP address.
	ApiserverHostIP string `structs:"apiserver_host_ip,omitempty" yaml:"apiserver_host_ip,omitempty" json:"apiserver_host_ip,omitempty"`

	// Kubernetes API server port.
	ApiserverHostPort string `structs:"apiserver_host_port,omitempty" yaml:"apiserver_host_port,omitempty" json:"apiserver_host_port,omitempty"`

	// Full path of the Token file to use for authenticating with the k8s API
	// server.
	ClientTokenFile string `structs:"client_token_file,omitempty" yaml:"client_token_file,omitempty" json:"client_token_file,omitempty"`

	// Full path of the client certificate file to use for authenticating with
	// the k8s API server. It must be specified together with
	// "client_private_key_file".
	ClientCertFile string `structs:"client_cert_file,omitempty" yaml:"client_cert_file,omitempty" json:"client_cert_file,omitempty"`

	// Full path of the client private key file to use for authenticating with
	// the k8s API server. It must be specified together with

	ClientPrivateKeyFile string `structs:"client_private_key_file,omitempty" yaml:"client_private_key_file,omitempty" json:"client_private_key_file,omitempty"`

	// Specify a CA bundle file to use in verifying the k8s API server
	// certificate.
	CaFile string `structs:"ca_file,omitempty" yaml:"ca_file,omitempty" json:"ca_file,omitempty"`

	// Specify whether ingress controllers are expected to be deployed in
	// hostnework mode or as regular pods externally accessed via NAT
	// Choices: hostnetwork nat
	IngressMode string `structs:"ingress_mode,omitempty" yaml:"ingress_mode,omitempty" json:"ingress_mode,omitempty"`

	// Log level for the kubernetes adaptor
	// Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL
	Loglevel string `structs:"loglevel,omitempty" yaml:"loglevel,omitempty" json:"loglevel,omitempty"`

	// The default HTTP ingress port

	HTTPIngressPort *int `structs:"http_ingress_port,omitempty" yaml:"http_ingress_port,omitempty" json:"http_ingress_port,omitempty"`

	// The default HTTPS ingress port
	HTTPSIngressPort *int `structs:"https_ingress_port,omitempty" yaml:"https_ingress_port,omitempty" json:"https_ingress_port,omitempty"`

	// Specify thread pool size to process resource events
	ResourceWatcherThreadPoolSize *int `structs:"resource_watcher_thread_pool_size,omitempty" yaml:"resource_watcher_thread_pool_size,omitempty" json:"resource_watcher_thread_pool_size,omitempty"`

	// User specified IP address for HTTP and HTTPS ingresses
	// nolint: golint, stylecheck
	HttpAndHttpsIngressIp string `structs:"http_and_https_ingress_ip,omitempty" yaml:"http_and_https_ingress_ip,omitempty" json:"http_and_https_ingress_ip,omitempty"`

	// Set this to True to enable NCP to create segment port for VM through
	// NsxNetworkInterface CRD.
	EnableNsxNetifCrd *bool `structs:"enable_nsx_netif_crd,omitempty" yaml:"enable_nsx_netif_crd,omitempty" json:"enable_nsx_netif_crd,omitempty"`

	// Option to set the type of baseline cluster policy. ALLOW_CLUSTER creates
	// an explicit baseline policy to allow any pod to communicate any other pod
	// within the cluster. ALLOW_NAMESPACE creates an explicit baseline policy
	// to allow pods within the same namespace to communicate with each other.
	// By default, no baseline rule will be created and the cluster will assume
	// the default behavior as specified by the backend.
	// Choices: <None> allow_cluster allow_namespace
	BaselinePolicyType string `structs:"baseline_policy_type,omitempty" yaml:"baseline_policy_type,omitempty" json:"baseline_policy_type,omitempty"`
}

type NsxNodeAgent struct {

	// The log level of NSX RPC library
	// Choices: NOTSET DEBUG INFO WARNING ERROR CRITICAL
	LogLevel string `structs:"nsxrpc_loglevel,omitempty" yaml:"log_level,omitempty" json:"log_level,omitempty"`

	// OVS bridge name
	OvsBridge string `structs:"ovs_bridge,omitempty" yaml:"ovs_bridge,omitempty" json:"ovs_bridge,omitempty"`

	// The OVS uplink OpenFlow port where to apply the NAT rules to.
	OvsUplinkPort string `structs:"ovs_uplink_port,omit_empty" yaml:"ovs_uplink_port,omitempty" json:"ovs_uplink_port,omitempty"`

	// The time in seconds for nsx_node_agent to wait CIF config from HyperBus
	// before returning to CNI
	ConfigRetryTimeout *int `structs:"config_retry_timeout,omitempty" yaml:"config_retry_timeout,omitempty" json:"config_retry_timeout,omitempty"`

	// The time in seconds for nsx_node_agent to backoff before re-using an
	// existing cached CIF to serve CNI request. Must be less than config_retry_timeout.
	ConfigReuseBackoffTime *int `structs:"config_reuse_backoff_time,omitempty" yaml:"config_reuse_backoff_time,omitempty" json:"config_reuse_backoff_time,omitempty"`
}
