package harbor

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	log "github.com/sirupsen/logrus"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

type HarborClient struct {
	sling *sling.Sling
	url   string
}

func NewHarborClient(p *platform.Platform) *HarborClient {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	return &HarborClient{
		url: p.Harbor.URL,
		sling: sling.New().Client(client).Base(p.Harbor.URL).
			SetBasicAuth("admin", p.Harbor.AdminPassword).
			Set("accept", "application/json").
			Set("content-type", "application/json"),
	}
}

func (harbor *HarborClient) ListReplicationPolicies() (policies []ReplicationPolicy, customErr error) {
	_, err := harbor.sling.New().
		Get("api/replication/policies").
		Receive(&policies, &customErr)
	if err == nil {
		err = customErr
	}
	return policies, err
}

func (harbor *HarborClient) TriggerReplication(id int) (*Replication, error) {
	req := Replication{PolicyID: id}
	r, err := harbor.sling.New().
		BodyJSON(&req).
		Post("api/replication/executions").
		Receive(nil, nil)
	if err != nil {
		return nil, err
	}
	_, err = harbor.sling.New().Get(r.Header["Location"][0]).Receive(&req, &req)
	return &req, err
}

func (harbor *HarborClient) UpdateSettings(settings types.HarborSettings) error {
	data, _ := json.Marshal(settings)
	log.Tracef("Harbor settings: \n%s\n", console.StripSecrets(string(data)))
	log.Infof("Updating harbor using: %s \n", harbor.url)
	r, err := harbor.sling.New().
		Put("api/configurations").
		BodyJSON(&settings).
		Receive(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update harbor settings: %v", err)
	}
	log.Debugf("Updated settings: %s:\n%+v", r.Status, text.SafeRead(r.Body))
	return nil
}

func (harbor *HarborClient) ListMembers(project string) ([]ProjectMember, error) {
	return nil, nil
}

type HarborProject struct {
}

type Replication struct {
	ID         int    `json:"id,omitempty"`
	PolicyID   int    `json:"policy_id,omitempty"`
	Status     string `json:"status,omitempty"`
	StatusText string `json:"status_text,omitempty"`
	Trigger    string `json:"trigger,omitempty"`
	Total      int    `json:"total,omitempty"`
	Failed     int    `json:"failed,omitempty"`
	Succeed    int    `json:"succeed,omitempty"`
	InProgress int    `json:"in_progress,omitempty"`
	Stopped    int    `json:"stopped,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
}

type ReplicationRegistry struct {
	ID         int    `json:"id,omitempty"`
	URL        string `json:"url,omitempty"`
	Name       string `json:"name,omitempty"`
	Credential struct {
		Type         string `json:"type,omitempty"`
		AccessKey    string `json:"access_key,omitempty"`
		AccessSecret string `json:"access_secret,omitempty"`
	} `json:"credential,omitempty"`
	Type         string `json:"type,omitempty"`
	Insecure     bool   `json:"insecure,omitempty"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status,omitempty"`
	CreationTime string `json:"creation_time,omitempty"`
	UpdateTime   string `json:"update_time,omitempty"`
}
type ReplicationPolicy struct {
	ID            int                 `json:"id,omitempty"`
	Name          string              `json:"name,omitempty"`
	Description   string              `json:"description,omitempty"`
	SrcRegistry   ReplicationRegistry `json:"src_registry,omitempty"`
	DestRegistry  ReplicationRegistry `json:"dest_registry,omitempty"`
	DestNamespace string              `json:"dest_namespace,omitempty"`
	Trigger       ReplicationTrigger  `json:"trigger,omitempty"`
	Filters       []ReplicationFilter `json:"filters,omitempty"`
	Deletion      bool                `json:"deletion,omitempty"`
	Override      bool                `json:"override,omitempty"`
	Enabled       bool                `json:"enabled,omitempty"`
	CreationTime  string              `json:"creation_time,omitempty"`
	UpdateTime    string              `json:"update_time,omitempty"`
}

type ReplicationTrigger struct {
	Type            string `json:"type,omitempty"`
	TriggerSettings struct {
		Cron string `json:"cron,omitempty"`
	} `json:"trigger_settings,omitempty"`
}

type ReplicationFilter struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type ProjectMember struct {
	ID         int    `json:"id,omitempty"`
	ProjectID  int    `json:"project_id,omitempty"`
	EntityName string `json:"entity_name,omitempty"`
	RoleName   string `json:"role_name,omitempty"`
	RoleID     int    `json:"role_id,omitempty"`
	EntityID   int    `json:"entity_id,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
}
