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

type Client struct {
	sling *sling.Sling
	url   string
}

func NewClient(p *platform.Platform) *Client {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	return &Client{
		url: p.Harbor.URL,
		sling: sling.New().Client(client).Base(p.Harbor.URL).
			SetBasicAuth("admin", p.Harbor.AdminPassword).
			Set("accept", "application/json").
			Set("content-type", "application/json"),
	}
}

func (harbor *Client) ListReplicationPolicies() (policies []ReplicationPolicy, customErr error) {
	resp, err := harbor.sling.New().
		Get("api/replication/policies").
		Receive(&policies, &customErr)
	if err == nil {
		err = customErr
	}
	defer resp.Body.Close()
	return policies, err
}

func (harbor *Client) TriggerReplication(id int) (*Replication, error) {
	req := Replication{PolicyID: id}
	r, err := harbor.sling.New().
		BodyJSON(&req).
		Post("api/replication/executions").
		Receive(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("TriggerReplication: failed to trigger replication: %v", err)
	}
	defer r.Body.Close()

	resp, err := harbor.sling.New().Get(r.Header["Location"][0]).Receive(&req, &req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &req, nil
}

func (harbor *Client) UpdateSettings(settings types.HarborSettings) error {
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
	defer r.Body.Close()
	log.Debugf("Updated settings: %s:\n%+v", r.Status, text.SafeRead(r.Body))
	return nil
}

func (harbor *Client) ListMembers(project string) ([]ProjectMember, error) {
	return nil, nil
}

type Project struct {
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
