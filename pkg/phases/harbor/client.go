package harbor

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/k8s/proxy"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	logger.Logger
	sling  *sling.Sling
	client *http.Client
	url    string
}

func NewClient(p *platform.Platform) (*Client, error) {
	clientset, err := p.GetClientset()
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(Namespace).List(metav1.ListOptions{
		LabelSelector: "app=harbor,component=core",
	})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no running harbor-core pods")
	}
	dialer, _ := p.GetProxyDialer(proxy.Proxy{
		Namespace:    Namespace,
		Kind:         "pods",
		ResourceName: pods.Items[0].Name,
		Port:         8080,
	})
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	return &Client{
		Logger: p.Logger,
		client: client,
		url:    p.Harbor.URL,
		sling: sling.New().Client(client).Base("http://harbor-core").
			SetBasicAuth("admin", p.Harbor.AdminPassword).
			Set("accept", "application/json").
			Set("content-type", "application/json"),
	}, nil
}

func (harbor *Client) GetStatus() (*Status, error) {
	resp, err := harbor.client.Get("http://harbor-core/api/health")
	if err != nil {
		return nil, err
	}
	status := Status{}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &status); err != nil {
		harbor.Errorf("Failed to unmarshall :%v", err)
	}
	return &status, nil
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
	harbor.Tracef("Harbor settings: \n%s\n", console.StripSecrets(string(data)))
	harbor.Infof("Updating harbor using: %s \n", harbor.url)
	r, err := harbor.sling.New().
		Put("api/configurations").
		BodyJSON(&settings).
		Receive(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update harbor settings: %v", err)
	}
	defer r.Body.Close()
	harbor.Debugf("Updated settings: %s:\n%+v", r.Status, text.SafeRead(r.Body))
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

type Status struct {
	Status     string `json:"status"`
	Components []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"components"`
}
