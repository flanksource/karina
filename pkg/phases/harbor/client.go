package harbor

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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

type IngressClient struct {
	logger.Logger
	sling  *sling.Sling
	client *http.Client
	url    string
	host   string
	base   string
}

// NewClient creates a new harbor client using proxy dialer
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

// NewIngressClient creates a new harbor client using the harbor ingress
func NewIngressClient(p *platform.Platform) (*IngressClient, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	harborDomain := fmt.Sprintf("harbor.%s", p.Domain)

	harborIP, err := dnsLookup(harborDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup harbor ingress IP")
	}

	harborURL := "https://" + harborIP

	var base string
	if strings.HasPrefix(p.Harbor.Version, "v1") {
		base = "/api"
	} else {
		base = "/api/v2.0"
	}
	client := &http.Client{Transport: tr}
	return &IngressClient{
		Logger: p.Logger,
		client: client,
		base:   base,
		url:    p.Harbor.URL,
		host:   harborDomain,
		sling: sling.New().Client(client).Base(harborURL+base).
			Set("Host", harborDomain).
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

func (harbor *IngressClient) ListProjects() ([]Project, error) {
	allProjects := []Project{}
	page := 1

	for {
		projects := []Project{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("https://%s/%s/projects?page=%d&per_page=%d", harbor.host, harbor.base, page, perPage)).
			Set("Host", harbor.host).
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

		allProjects = append(allProjects, projects...)

		if len(projects) == 0 {
			break
		}
		page++
	}

	return allProjects, nil
}

func (harbor *IngressClient) ListImages(project string) ([]Image, error) {
	allImages := []Image{}
	page := 1

	for {
		images := []Image{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("https://%s/%s/projects/%s/repositories?page=%d&perPage=%d", harbor.host, harbor.base, project, page, perPage)).
			Set("Host", harbor.host).
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

		allImages = append(allImages, images...)

		if len(images) == 0 {
			break
		}
		page++
	}
	return allImages, nil
}

func (harbor *IngressClient) ListArtifacts(project string, image string) ([]Artifact, error) {
	allArtifacts := []Artifact{}
	page := 1

	imageEncoded := url.QueryEscape(image)

	for {
		artifacts := []Artifact{}
		er := ErrorResponse{}
		r, err := harbor.sling.New().
			Get(fmt.Sprintf("https://%s/%s/projects/%s/repositories/%s/artifacts?page=%d&per_page=%d", harbor.host, harbor.base, project, imageEncoded, page, perPage)).
			Set("Host", harbor.host).
			Receive(&artifacts, &er)
		if r != nil {
			r.Body.Close()
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list tags")
		}

		if len(er.Errors) > 0 {
			return nil, errors.Errorf("received error code from server: %s", er.String())
		}

		allArtifacts = append(allArtifacts, artifacts...)

		if len(artifacts) == 0 {
			break
		}
		page++
	}
	return allArtifacts, nil
}

func (harbor *IngressClient) GetManifest(project, image, tag string) (*Manifest, error) {
	manifest := &Manifest{}
	er := ErrorResponse{}
	// fmt.Printf("url: https://%s/v2/%s/%s/manifests/%s\n", harbor.host, project, image, tag)
	r, err := harbor.sling.New().
		Get(fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s", harbor.host, project, image, tag)).
		Set("Host", harbor.host).
		Receive(&manifest, &er)
	if r != nil {
		r.Body.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest")
	}

	if len(er.Errors) > 0 {
		return nil, errors.Errorf("received error code from server: %s", er.String())
	}

	return manifest, nil
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

type Artifact struct {
	ID                int           `json:"id,omitempty"`
	Type              string        `json:"type,omitempty"`
	MediaType         string        `json:"media_type,omitempty"`
	ManifestMediaType string        `json:"manifest_media_type,omitempty"`
	ProjectID         int           `json:"project_id,omitempty"`
	RepositoryID      int           `json:"repository_id,omitempty"`
	Digest            string        `json:"digest,omitempty"`
	Size              int           `json:"size,omitempty"`
	Tags              []ArtifactTag `json:"tags"`
}

type ArtifactTag struct {
	Name string `json:"name"`
}

type Tag struct {
	Name           string `json:"-"`
	ProjectName    string `json:"-"`
	RepositoryName string `json:"-"`
	Digest         string `json:"-"`
}

type Manifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        ManifestLayer   `json:"config"`
	Layers        []ManifestLayer `json:"layers"`
}

type ManifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
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

func dnsLookup(hostname string) (string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", errors.Wrapf(err, "failed to lookup %s", hostname)
	}

	if len(addrs) == 0 {
		return "", errors.Errorf("lookup %s did not return any address", hostname)
	}

	return addrs[0], nil
}
