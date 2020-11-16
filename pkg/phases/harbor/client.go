package harbor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/sling"
	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/text"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons/proxy"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const perPage = 50

type Client struct {
	logger.Logger
	sling  *sling.Sling
	client *http.Client
	url    string
	base   string
}

func NewClient(p *platform.Platform) (*Client, error) {
	clientset, err := p.GetClientset()
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{
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
		Port:         8443,
	})
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var base string
	if strings.HasPrefix(p.Harbor.Version, "v1") {
		base = "/api"
	} else {
		base = "/api/v2.0"
	}
	client := &http.Client{Transport: tr}
	return &Client{
		Logger: p.Logger,
		client: client,
		base:   base,
		url:    p.Harbor.URL,
		sling: sling.New().Client(client).Base("https://harbor-core"+base).
			SetBasicAuth("admin", p.Harbor.AdminPassword).
			Set("accept", "application/json").
			Set("content-type", "application/json"),
	}, nil
}

func (harbor *Client) GetStatus() (*Status, error) {
	resp, err := harbor.client.Get("https://harbor-core" + harbor.base + "/health")
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
		Get("replication/policies").
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
		Post("replication/executions").
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
	harbor.Infof("Updating harbor using: %s \n", harbor.base)
	r, err := harbor.sling.New().
		Put(harbor.base+"/configurations").
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

func (harbor *Client) ListImages(project string) (images []Image, customError error) {
	allImages := []Image{}
	page := 1

	for {
		images := []Image{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("%s/projects/%s/repositories?page=%d&perPage=%d", harbor.base, project, page, perPage)).
			Receive(&images, &er)
		if r != nil {
			r.Body.Close()
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list images")
		}

		if len(er.Errors) > 0 {
			return nil, errors.Errorf("received error code from server: %s", er.String())
		}

		harbor.Debugf("images for project %s page %d %d results", project, page, len(images))

		allImages = append(allImages, images...)

		if len(images) == 0 {
			break
		}
		page++
	}
	return allImages, nil
}

func (harbor *Client) ListProjects() ([]Project, error) {
	allProjects := []Project{}
	page := 1

	for {
		projects := []Project{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("%s/projects?page=%d&per_page=%d", harbor.base, page, perPage)).
			Receive(&projects, &er)
		if r != nil {
			r.Body.Close()
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list projects")
		}

		if len(er.Errors) > 0 {
			return nil, errors.Errorf("received error code from server: %s", er.String())
		}

		harbor.Debugf("projects page %d  %d results", page, len(projects))

		allProjects = append(allProjects, projects...)

		if len(projects) == 0 {
			break
		}
		page++
	}

	return allProjects, nil
}

func (harbor *Client) ListTags(project string, image string) (tags []Tag, customError error) {
	allTags := []Tag{}
	page := 1

	imageEncoded := url.QueryEscape(image)

	for {
		tags := []Tag{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts?page=%d&per_page=%d", harbor.base, project, imageEncoded, page, perPage)).
			Receive(&tags, &er)
		if r != nil {
			r.Body.Close()
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list tags")
		}

		if len(er.Errors) > 0 {
			return nil, errors.Errorf("received error code from server: %s", er.String())
		}

		harbor.Debugf("project %s image %s tags page %d  %d results", project, image, page, len(tags))

		allTags = append(allTags, tags...)

		if len(tags) == 0 {
			break
		}
		page++
	}
	return allTags, nil
}

type Project struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
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

type Image struct {
	ID            int    `json:"id,omitempty"`
	ProjectID     int    `json:"project_id,omitempty"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	ArtifactCount int    `json:"artifact_count,omitempty"`
	PullCount     int    `json:"pull_count,omitempty"`

	ProjectName string `json:"-"`
}

type Tag struct {
	ID                int    `json:"id,omitempty"`
	Type              string `json:"type,omitempty"`
	MediaType         string `json:"media_type,omitempty"`
	ManifestMediaType string `json:"manifest_media_type,omitempty"`
	ProjectID         int    `json:"project_id,omitempty"`
	RepositoryID      int    `json:"repository_id,omitempty"`
	Digest            string `json:"digest,omitempty"`
	Size              int    `json:"size,omitempty"`

	ProjectName    string `json:"-"`
	RepositoryName string `json:"-"`
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func (e *ErrorResponse) String() string {
	j, _ := json.Marshal(e)
	return string(j)
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
