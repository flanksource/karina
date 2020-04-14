package elasticsearch

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Info represents the response from /
type Info struct {
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number string `json:"number"`
	} `json:"version"`
}

// Health is the health of the cluster as returned by the health API.
type HealthString string

// Possible traffic light states Elasticsearch health can have.
const (
	RedHealth     HealthString = "red"
	YellowHealth  HealthString = "yellow"
	GreenHealth   HealthString = "green"
	UnknownHealth HealthString = "unknown"
)

// Health represents the response from _cluster/health
type Health struct {
	ClusterName                 string       `json:"cluster_name"`
	Status                      HealthString `json:"status"`
	TimedOut                    bool         `json:"timed_out"`
	NumberOfNodes               int          `json:"number_of_nodes"`
	NumberOfDataNodes           int          `json:"number_of_data_nodes"`
	ActivePrimaryShards         int          `json:"active_primary_shards"`
	ActiveShards                int          `json:"active_shards"`
	RelocatingShards            int          `json:"relocating_shards"`
	InitializingShards          int          `json:"initializing_shards"`
	UnassignedShards            int          `json:"unassigned_shards"`
	DelayedUnassignedShards     int          `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int          `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int          `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int          `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float32      `json:"active_shards_percent_as_number"`
}

func (h Health) String() string {
	return fmt.Sprintf("%s nodes=%d shards=%d replicas=%d relocating=%d unassigned=%d initializing=%d", h.Status, h.NumberOfNodes, h.ActivePrimaryShards, h.ActiveShards-h.ActivePrimaryShards, h.RelocatingShards, h.UnassignedShards, h.InitializingShards)
}

type ShardState string

// These are possible shard states
const (
	STARTED      ShardState = "STARTED"
	INITIALIZING ShardState = "INITIALIZING"
	RELOCATING   ShardState = "RELOCATING"
	UNASSIGNED   ShardState = "UNASSIGNED"
)

type ShardType string

const (
	Primary ShardType = "p"
	Replica ShardType = "r"
)

// Nodes partially models the response from a request to /_nodes
type Nodes struct {
	Nodes map[string]Node `json:"nodes"`
}

func (n Nodes) Names() []string {
	names := make([]string, 0, len(n.Nodes))
	for _, node := range n.Nodes {
		names = append(names, node.Name)
	}
	return names
}

// Node partially models an Elasticsearch node retrieved from /_nodes
type Node struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Roles   []string `json:"roles"`
	JVM     struct {
		StartTimeInMillis int64 `json:"start_time_in_millis"`
		Mem               struct {
			HeapMaxInBytes int `json:"heap_max_in_bytes"`
		} `json:"mem"`
	} `json:"jvm"`
}

// NodesStats partially models the response from a request to /_nodes/stats
type NodesStats struct {
	Nodes map[string]NodeStats `json:"nodes"`
}

// NodeStats partially models an Elasticsearch node retrieved from /_nodes/stats
type NodeStats struct {
	Name string `json:"name"`
	OS   struct {
		CGroup struct {
			Memory struct {
				LimitInBytes string `json:"limit_in_bytes"`
			} `json:"memory"`
		} `json:"cgroup"`
	} `json:"os"`
}

// ClusterStateNode represents an element in the `node` structure in
// Elasticsearch cluster state.
type ClusterStateNode struct {
	Name             string `json:"name"`
	EphemeralID      string `json:"ephemeral_id"`
	TransportAddress string `json:"transport_address"`
	Attributes       struct {
		MlMachineMemory string `json:"ml.machine_memory"`
		MlMaxOpenJobs   string `json:"ml.max_open_jobs"`
		XpackInstalled  string `json:"xpack.installed"`
		MlEnabled       string `json:"ml.enabled"`
	} `json:"attributes"`
}

// Shards contains the shards in the Elasticsearch cluster
type Shards []Shard

// Shard partially models Elasticsearch cluster shard.
type Shard struct {
	Index    string     `json:"index"`
	Shard    string     `json:"shard"`
	State    ShardState `json:"state"`
	NodeName string     `json:"node"`
	Type     ShardType  `json:"prirep"`
}

type RoutingTable struct {
	Indices map[string]Shards `json:"indices"`
}

// GetShardsByNode returns shards by node.
// The result is a map with the name of the nodes as keys and the list of shards on the nodes as values.
func (s Shards) GetShardsByNode() map[string]Shards {
	result := make(map[string]Shards)
	for _, shard := range s {
		// Unassigned shards are ignored
		if len(shard.NodeName) > 0 {
			result[shard.NodeName] = append(result[shard.NodeName], shard)
		}
	}
	return result
}

// Strip extra information from the nodeName field
// eg. "cluster-node-2 -> 10.56.2.33 8DqGuLtrSNyMfE2EfKNDgg" becomes "cluster-node-2"
// see https://github.com/elastic/cloud-on-k8s/issues/1796
func (s *Shards) UnmarshalJSON(data []byte) error {
	type Alias Shards
	aux := (*Alias)(s)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for i, shard := range *aux {
		if idx := strings.IndexByte(shard.NodeName, ' '); idx >= 0 {
			(*s)[i].NodeName = (*s)[i].NodeName[:idx]
		}
	}
	return nil
}

// IsRelocating is true if the shard is relocating to another node.
func (s Shard) IsRelocating() bool {
	return s.State == RELOCATING
}

// IsStarted is true if the shard is started on its current node.
func (s Shard) IsStarted() bool {
	return s.State == STARTED
}

// IsInitializing is true if the shard is currently initializing on the node.
func (s Shard) IsInitializing() bool {
	return s.State == INITIALIZING
}

// IsReplica is true if the shard is a replica.
func (s Shard) IsReplica() bool {
	return s.Type == Replica
}

// IsPrimary is true if the shard is a primary shard.
func (s Shard) IsPrimary() bool {
	return s.Type == Primary
}

// Key is a composite key of index name and shard number that identifies all
// copies of a shard across nodes.
func (s Shard) Key() string {
	return concat(s.Index, "/", s.Shard)
}

// Concat joins arguments to form a concatenated string. Uses strings.Builder
// To concatenate in the most efficient manner.
func concat(args ...string) string {
	var result strings.Builder
	for _, arg := range args {
		// it's safe to ignore the result here as strings.Builder cannot error on result.WriteString
		result.WriteString(arg) // #nosec G104
	}
	return result.String()
}

// AllocationSettings model a subset of the supported attributes for dynamic Elasticsearch cluster settings.
type AllocationSettings struct {
	Cluster ClusterRoutingSettings `json:"cluster,omitempty"`
} // TODO awareness settings

type ClusterRoutingSettings struct {
	Routing RoutingSettings `json:"routing,omitempty"`
}

type RoutingSettings struct {
	Allocation RoutingAllocationSettings `json:"allocation,omitempty"`
}

type RoutingAllocationSettings struct {
	Exclude AllocationExclude `json:"exclude,omitempty"`
	Enable  string            `json:"enable,omitempty"`
}

type AllocationExclude struct {
	Name string `json:"_name,omitempty"`
}

func (s AllocationSettings) IsShardsAllocationEnabled() bool {
	enable := s.Cluster.Routing.Allocation.Enable
	return enable == "" || enable == "all"
}

// ClusterRoutingAllocation models a subset of transient allocation settings for an Elasticsearch cluster.
type ClusterRoutingAllocation struct {
	Transient AllocationSettings `json:"transient,omitempty"`
}

// DiscoveryZen set minimum number of master eligible nodes that must be visible to form a cluster.
type DiscoveryZen struct {
	MinimumMasterNodes int `json:"discovery.zen.minimum_master_nodes"`
}

// DiscoveryZenSettings are cluster settings related to the zen discovery mechanism.
type DiscoveryZenSettings struct {
	Transient  DiscoveryZen `json:"transient"`
	Persistent DiscoveryZen `json:"persistent"`
}

// ErrorResponse is a Elasticsearch error response.
type ErrorResponse struct {
	Status int `json:"status"`
	Error  struct {
		CausedBy struct {
			Reason string `json:"reason"`
			Type   string `json:"type"`
		} `json:"caused_by"`
		Reason    string `json:"reason"`
		Type      string `json:"type"`
		RootCause []struct {
			Reason string `json:"reason"`
			Type   string `json:"type"`
		} `json:"root_cause"`
	} `json:"error"`
}

// RemoteClustersSettings is used to build a request to update remote clusters.
type RemoteClustersSettings struct {
	PersistentSettings *SettingsGroup `json:"persistent,omitempty"`
}

// SettingsGroup is a group of persistent settings.
type SettingsGroup struct {
	Cluster RemoteClusters `json:"cluster,omitempty"`
}

// RemoteClusters models the configuration of the remote clusters.
type RemoteClusters struct {
	RemoteClusters map[string]RemoteCluster `json:"remote,omitempty"`
}

// RemoteClusterSeeds is the set of seeds to use in a remote cluster setting.
type RemoteCluster struct {
	Seeds []string `json:"seeds"`
}

// Hit represents a single search hit.
type Hit struct {
	Index  string                 `json:"_index"`
	Type   string                 `json:"_type"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// Hits are the collections of search hits.
type Hits struct {
	Total json.RawMessage // model when needed
	Hits  []Hit           `json:"hits"`
}

// SearchResults are the results returned from a _search.
type SearchResults struct {
	Took   int
	Hits   Hits                       `json:"hits"`
	Shards json.RawMessage            // model when needed
	Aggs   map[string]json.RawMessage // model when needed
}
