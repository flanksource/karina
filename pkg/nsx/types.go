package nsx

import "github.com/vmware/go-vmware-nsxt/loadbalancer"

const (
	TCPProtocol = "TCP"
)

type LBService struct {
	PolicyBase           `json:",inline"`
	Enabled              bool   `json:"enabled"`
	RelaxScaleValidation bool   `json:"relax_scale_validation"`
	Size                 string `json:"size"`
	ErrorLogLevel        string `json:"error_log_level"`
	AccessLogEnabled     bool   `json:"access_log_enabled"`
}

type LBVirtualServer struct {
	PolicyBase               `json:",inline"`
	Enabled                  bool     `json:"enabled"`
	IPAddress                string   `json:"ip_address"`
	Ports                    []string `json:"ports"`
	AccessLogEnabled         bool     `json:"access_log_enabled"`
	LbPersistenceProfilePath string   `json:"lb_persistence_profile_path"`
	LbServicePath            string   `json:"lb_service_path"`
	PoolPath                 string   `json:"pool_path"`
	ApplicationProfilePath   string   `json:"application_profile_path"`
	LogSignificantEventOnly  bool     `json:"log_significant_event_only"`
}

type LoadBalancerPool struct {
	Members []LoadBalancerPoolMember `json:"members,omitempty"`
}

type LoadBalancerPoolMember struct {
	AdminState string `json:"admin_state"`
	IPAddress  string `json:"ip_address"`
}

type LoadBalancerPoolResults struct {
	ResultCount int                `json:"result_count"`
	Results     []LoadBalancerPool `json:"results"`
}

type PolicyLoadBalancerPool struct {
	PolicyBase  `json:",inline"`
	Algorithm   string `json:"algorithm"`
	MemberGroup struct {
		GroupPath        string `json:"group_path"`
		Port             int    `json:"port"`
		IPRevisionFilter string `json:"ip_revision_filter"`
	} `json:"member_group"`
	ActiveMonitorPaths []string `json:"active_monitor_paths"`
	SnatTranslation    struct {
		Type string `json:"type"`
	} `json:"snat_translation"`
	TCPMultiplexingEnabled bool `json:"tcp_multiplexing_enabled"`
	TCPMultiplexingNumber  int  `json:"tcp_multiplexing_number"`
	MinActiveMembers       int  `json:"min_active_members"`
}
type LBTcpMonitorProfile struct {
	PolicyBase  `json:",inline"`
	MonitorPort int `json:"monitor_port"`
	Interval    int `json:"interval"`
	Timeout     int `json:"timeout"`
	RiseCount   int `json:"rise_count"`
	FallCount   int `json:"fall_count"`
}

type PolicyBase struct {
	ResourceType string `json:"resource_type"`
	ID           string `json:"id"`
	DisplayName  string `json:"display_name"`
	Tags         []struct {
		Scope string `json:"scope"`
		Tag   string `json:"tag"`
	} `json:"tags,omitempty"`
	Path            string `json:"path,omitempty"`
	RelativePath    string `json:"relative_path,omitempty"`
	ParentPath      string `json:"parent_path,omitempty"`
	UniqueID        string `json:"unique_id,omitempty"`
	MarkedForDelete bool   `json:"marked_for_delete,omitempty"`
	Overridden      bool   `json:"overridden,omitempty"`
	AuditBase       `json:",inline"`
}

type AuditBase struct {
	CreateTime       int64  `json:"_create_time"`
	LastModifiedUser string `json:"_last_modified_user"`
	LastModifiedTime int64  `json:"_last_modified_time"`
	SystemOwned      bool   `json:"_system_owned"`
	CreateUser       string `json:"_create_user"`
	Protection       string `json:"_protection"`
	Revision         int    `json:"_revision"`
}

type SegmentPort struct {
	PolicyBase `json:",inline"`
	Attachment struct {
		ID           string `json:"id"`
		TrafficTag   int    `json:"traffic_tag"`
		HyperbusMode string `json:"hyperbus_mode"`
	} `json:"attachment"`
	AdminState string `json:"admin_state"`
}

type LoadBalancerOptions struct {
	Name       string
	IPPool     string
	Protocol   string
	Ports      []string
	Tier0      string
	MemberTags map[string]string
}

type virtualServersList struct {
	Results []loadbalancer.LbVirtualServer `json:"results,omitempty"`
}

type virtualPoolList struct {
	Results []loadbalancer.LbPool `json:"results,omitempty"`
}
