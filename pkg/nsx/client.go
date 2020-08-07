package nsx

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
	"github.com/vmware/go-vmware-nsxt/manager"
	"k8s.io/apimachinery/pkg/util/wait"
)

// nolint: golint
type NSXClient struct {
	api *nsxt.APIClient
	cfg *nsxt.Configuration
	logger.Logger
	Username, Password, Host string
	RemoteAuth               bool
}

func (c *NSXClient) Init() error {
	if c.api != nil {
		return nil
	}

	retriesConfig := nsxt.ClientRetriesConfiguration{
		MaxRetries:      3,
		RetryMinDelay:   1,
		RetryMaxDelay:   5,
		RetryOnStatuses: []int{429, 503},
	}

	cfg := nsxt.Configuration{
		UserName:   c.Username,
		Password:   c.Password,
		BasePath:   "/api/v1",
		Host:       c.Host,
		Scheme:     "https",
		UserAgent:  "karina",
		RemoteAuth: c.RemoteAuth,
		DefaultHeader: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)),
		},
		Insecure:             true,
		RetriesConfiguration: retriesConfig,
	}

	client, err := nsxt.NewAPIClient(&cfg)
	if err != nil {
		return fmt.Errorf("Init: Failed to get nsxt API client: %v", err)
	}
	c.api = client
	c.cfg = &cfg
	return nil
}

type NSXError struct {
	HTTPStatus   string     `json:"httpStatus"`
	ErrorCode    int        `json:"error_code"`
	ModuleName   string     `json:"module_name"`
	ErrorMessage string     `json:"error_message"`
	Related      []NSXError `json:"related_errors"`
}

func getError(body []byte, code int) error {
	if code >= 200 && code <= 299 {
		return nil
	}

	err := NSXError{}
	if unmarshallErr := json.Unmarshal(body, &err); unmarshallErr == nil {
		if len(err.Related) > 0 {
			err = err.Related[0]
		}
		if err.ErrorCode == code {
			return fmt.Errorf("%d: %s", code, err.ErrorMessage)

		} else {
			return fmt.Errorf("%d: %d: %s", code, err.ErrorCode, err.ErrorMessage)

		}
	}
	return fmt.Errorf("%d: %s", code, string(body))
}

func (c *NSXClient) GET(path string) ([]byte, int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	url := fmt.Sprintf("%s://%s%s%s", c.cfg.Scheme, c.cfg.Host, c.cfg.BasePath, path)
	if strings.HasPrefix(path, "/policy/api") {
		url = fmt.Sprintf("%s://%s%s", c.cfg.Scheme, c.cfg.Host, path)
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range c.cfg.DefaultHeader {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	c.Logger.Tracef("GET: %s -> %d", url, resp.StatusCode)

	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("GET: Failed to read response: %v", err)
	}

	return body, resp.StatusCode, getError(body, resp.StatusCode)
}

func (c *NSXClient) POST(path string, body interface{}) ([]byte, int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	url := fmt.Sprintf("%s://%s%s%s", c.cfg.Scheme, c.cfg.Host, c.cfg.BasePath, path)
	if strings.HasPrefix(path, "/policy/api") {
		url = fmt.Sprintf("%s://%s%s", c.cfg.Scheme, c.cfg.Host, path)
	}

	req, _ := http.NewRequest("POST", url, nil)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")
	for k, v := range c.cfg.DefaultHeader {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	c.Logger.Debugf("POST: %s -> %d", url, resp.StatusCode)

	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("POST: Failed to read response: %v", err)
	}
	c.Logger.Tracef("POST: %s -> %d, body=%s resp=%s", url, resp.StatusCode, string(reqBody), string(respBody))
	return respBody, resp.StatusCode, getError(respBody, resp.StatusCode)
}

func (c *NSXClient) Ping() (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("need to called .Init() first")
	}
	props, resp, err := c.api.NsxComponentAdministrationApi.ReadNodeProperties(c.api.Context)
	if err != nil {
		return "", fmt.Errorf("error pinging: %v: resp: %v, req: %v", err, resp.Header, resp.Request.Header)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("failed to get version: %s", resp.Status)
	}
	return props.NodeVersion, nil
}

func errorString(resp *http.Response, err error) string {
	s := fmt.Sprintf("%s", err)
	if id, ok := resp.Header["X-Nsx-Requestid"]; ok {
		s += fmt.Sprintf("%s reqid=%s", s, id)
	}
	return s
}

func (c *NSXClient) GetLogicalPorts(ctx context.Context, vm string) ([]manager.LogicalPort, error) {
	var results []manager.LogicalPort
	vms, resp, err := c.api.FabricApi.ListVirtualMachines(c.api.Context, map[string]interface{}{"displayName": vm})
	if err != nil {
		return nil, fmt.Errorf("cannot get vifs for %s: %v", vm, errorString(resp, err))
	}

	if vms.ResultCount == 0 {
		return nil, fmt.Errorf("vm %s not found", vm)
	}
	vifs, resp, err := c.api.FabricApi.ListVifs(c.api.Context, map[string]interface{}{"ownerVmId": vms.Results[0].ExternalId})
	if err != nil {
		return nil, fmt.Errorf("cannot get vifs for %s: %v", vm, errorString(resp, err))
	}

	for _, vif := range vifs.Results {
		ports, resp, err := c.api.LogicalSwitchingApi.ListLogicalPorts(c.api.Context, map[string]interface{}{"attachmentId": vif.LportAttachmentId})

		if err != nil {
			return nil, fmt.Errorf("unable to get port %s: %s", vif.LportAttachmentId, errorString(resp, err))
		}
		results = append(results, ports.Results...)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no vifs/logical ports found for vm: %s, externalId: %s", vm, vms.Results[0].ExternalId)
	}
	return results, nil
}

func (c *NSXClient) TagLogicalPort(ctx context.Context, id string, tags map[string]string) error {
	port, resp, err := c.api.LogicalSwitchingApi.GetLogicalPort(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to get port %s: %s", id, errorString(resp, err))
	}

	for k, v := range tags {
		port.Tags = append(port.Tags, common.Tag{
			Scope: k,
			Tag:   v,
		})
	}

	c.Tracef("[%s/%s] tagging: %v", port.Id, port.Attachment.Id, port.Tags)
	_, resp, err = c.api.LogicalSwitchingApi.UpdateLogicalPort(context.TODO(), port.Id, port)
	if err != nil {
		return fmt.Errorf("unable to update port %s: %s", port.Id, errorString(resp, err))
	}
	return nil
}

// nolint: structcheck, unused, golint, stylecheck
type LoadBalancer struct {
	client *NSXClient
	name   string
	Id     string
}

// nolint: golint, stylecheck
type NSGroup struct {
	client   *NSXClient
	Name, Id string
}

func (group *NSGroup) Add(ips ...string) error {
	return nil
}

func (group *NSGroup) Remove(ips ...string) error {
	return nil
}
func (group *NSGroup) List() ([]string, error) {
	return nil, nil
}

func (c *NSXClient) CreateOrUpdateNSGroup(name string, targetType string, tags map[string]string) (*NSGroup, error) {
	ctx := c.api.Context
	var criteria []manager.NsGroupTagExpression
	for k, v := range tags {
		criteria = append(criteria, manager.NsGroupTagExpression{
			ResourceType: "NSGroupTagExpression",
			TargetType:   targetType,
			Tag:          v,
			Scope:        k,
		})
	}

	// nolint: bodyclose
	_, resp, err := c.api.GroupingObjectsApi.ReadNSGroup(ctx, name, map[string]interface{}{})
	if err != nil || resp != nil && resp.StatusCode == http.StatusNotFound {
		// nolint: bodyclose
		group, resp, err := c.api.GroupingObjectsApi.CreateNSGroup(c.api.Context, manager.NsGroup{
			Id:                 name,
			ResourceType:       "NSGroupTagExpression",
			MembershipCriteria: criteria,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create/update NSGroup %s: %s", name, errorString(resp, err))
		}
		return &NSGroup{
			client: c,
			Name:   group.DisplayName,
			Id:     group.Id,
		}, nil
	}
	// nolint: bodyclose
	group, resp, err := c.api.GroupingObjectsApi.UpdateNSGroup(ctx, name, manager.NsGroup{
		Id:                 name,
		ResourceType:       "NSGroupTagExpression",
		MembershipCriteria: criteria,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create/update NSGroup %s: %s", name, errorString(resp, err))
	}
	return &NSGroup{
		client: c,
		Name:   group.DisplayName,
		Id:     group.Id,
	}, nil
}

func (c *NSXClient) AllocateIP(pool string) (string, error) {
	addr, resp, err := c.api.PoolManagementApi.AllocateOrReleaseFromIpPool(c.api.Context, pool, manager.AllocationIpAddress{}, "ALLOCATE")
	if err != nil {
		return "", fmt.Errorf("unable to allocate IP from %s: %s", pool, errorString(resp, err))
	}
	c.Infof("Allocated IP: %s %s %s", addr.AllocationId, addr.Id, addr.DisplayName)
	return addr.AllocationId, nil
}

func (c *NSXClient) GetLoadBalancerPool(name string) (*loadbalancer.LbPool, error) {
	// ServicesApi.GetLoadBalancerPools() is not implemented
	body, _, err := c.GET("/loadbalancer/pools")
	if err != nil {
		return nil, fmt.Errorf("failed to list existing virtual servers: %v", err)
	}

	var pools virtualPoolList

	if err := json.Unmarshal(body, &pools); err != nil {
		return nil, fmt.Errorf("failed to unmarshall existing virtual pool list: %v", err)
	}

	for _, server := range pools.Results {
		if server.DisplayName == name {
			return &server, nil
		}
	}
	return nil, nil
}

func (c *NSXClient) GetLoadBalancer(name string) (*loadbalancer.LbVirtualServer, error) {
	// ServicesApi.GetVirtualServers() is not implemented
	body, _, err := c.GET("/loadbalancer/virtual-servers")
	if err != nil {
		return nil, fmt.Errorf("failed to list existing virtual servers: %v", err)
	}

	var virtualServers virtualServersList

	if err := json.Unmarshal(body, &virtualServers); err != nil {
		return nil, fmt.Errorf("failed to unmarshall existing virtual server list: %v", err)
	}

	for _, server := range virtualServers.Results {
		if server.DisplayName == name {
			return &server, nil
		}
	}
	return nil, nil
}

var backoffOptions = wait.Backoff{
	Duration: 500 * time.Millisecond,
	Factor:   2.0,
	Steps:    8,
}

func (c *NSXClient) DrainLoadBalancerMember(name, ip string) error {
	pool, err := c.GetLoadBalancerPool(name)
	if err != nil {
		return err
	}
	return wait.ExponentialBackoff(backoffOptions, func() (bool, error) {
		body, code, err := c.POST(fmt.Sprintf("/loadbalancer/pools/%s?action=UPDATE_MEMBERS", pool.Id), LoadBalancerPool{Members: []LoadBalancerPoolMember{
			{
				IPAddress:  ip,
				AdminState: "GRACEFUL_DISABLED",
			},
		}})

		if err != nil {
			c.Logger.Warnf("error while draining lb, retrying %s", err)
			return false, nil
		}

		if code != 200 {
			c.Logger.Warnf("non-ok status code while draining lb, retrying %d: %v", code, string(body))
			return false, nil
		}

		c.Logger.Infof("Removed %s from %s", ip, name)
		return true, nil
	})
}

// CreateLoadBalancer creates a new loadbalancer or returns the existing loadbalancer's IP
func (c *NSXClient) CreateLoadBalancer(opts LoadBalancerOptions) (string, bool, error) {
	ctx := c.api.Context
	api := c.api.ServicesApi
	routing := c.api.LogicalRoutingAndServicesApi

	existingServer, err := c.GetLoadBalancer(opts.Name)
	if err != nil {
		return "", false, err
	}
	if existingServer != nil {
		return existingServer.IpAddress, true, nil
	}

	t0, resp, err := routing.ReadLogicalRouter(ctx, opts.Tier0)
	if err != nil {
		return "", false, fmt.Errorf("failed to read T0 router %s: %s", opts.Tier0, errorString(resp, err))
	}

	t0Port, resp, err := routing.CreateLogicalRouterLinkPortOnTier0(ctx, manager.LogicalRouterLinkPortOnTier0{
		LogicalRouterId: t0.Id,
		DisplayName:     "lb-" + opts.Name + "-T1",
	})

	if err != nil {
		return "", false, fmt.Errorf("unable to create T0 Local router port %s: %s", opts.Name, errorString(resp, err))
	}

	t1, resp, err := routing.CreateLogicalRouter(ctx, manager.LogicalRouter{
		RouterType:    "TIER1",
		DisplayName:   "lb-" + opts.Name,
		EdgeClusterId: t0.EdgeClusterId,
	})
	if err != nil {
		return "", false, fmt.Errorf("unable to create T1 router %s: %s", opts.Name, errorString(resp, err))
	}

	_, resp, err = routing.UpdateAdvertisementConfig(ctx, t1.Id, manager.AdvertisementConfig{
		AdvertiseLbVip:    true,
		AdvertiseLbSnatIp: true,
		Enabled:           true,
	})
	if err != nil {
		return "", false, fmt.Errorf("unable to update advertisement config %s: %s", opts.Name, errorString(resp, err))
	}

	c.Infof("Created T1 router %s/%s", t1.DisplayName, t1.Id)

	_, resp, err = routing.CreateLogicalRouterLinkPortOnTier1(ctx, manager.LogicalRouterLinkPortOnTier1{
		LogicalRouterId: t1.Id,
		DisplayName:     t0.DisplayName + "-uplink",
		LinkedLogicalRouterPortId: &common.ResourceReference{
			TargetType: "LogicalPort",
			TargetId:   t0Port.Id,
		},
	})
	if err != nil {
		return "", false, fmt.Errorf("failed to link T1 (%s) to T0 (%s): %s", t1.Id, t0Port.Id, errorString(resp, err))
	}

	group, err := c.CreateOrUpdateNSGroup(opts.Name, "LogicalPort", opts.MemberTags)
	if err != nil {
		return "", false, err
	}
	var monitorID string
	if opts.Protocol == TCPProtocol {
		monitorID, err = c.GetOrCreateTCPHealthCheck(opts.Ports[0])
		if err != nil {
			return "", false, fmt.Errorf("unable to create tcp loadbalancer monitor: %v", err)
		}
	} else {
		monitorID, err = c.GetOrCreateHTTPHealthCheck(opts.Ports[0])
		if err != nil {
			return "", false, fmt.Errorf("unable to create http loadbalancer monitor: %v", err)
		}
	}
	pool, resp, err := api.CreateLoadBalancerPool(ctx, loadbalancer.LbPool{
		Id:               opts.Name,
		ActiveMonitorIds: []string{monitorID},
		SnatTranslation: &loadbalancer.LbSnatTranslation{
			Type_: "LbSnatAutoMap",
		},
		MemberGroup: &loadbalancer.PoolMemberGroup{
			GroupingObject: &common.ResourceReference{
				TargetType: "NSGroup",
				TargetId:   group.Id,
			},
		},
	})
	if err != nil {
		return "", false, fmt.Errorf("unable to create load balancer pool %s: %s", opts.Name, errorString(resp, err))
	}

	ip, err := c.AllocateIP(opts.IPPool)
	if err != nil {
		return "", false, fmt.Errorf("unable to allocate VIP %s: %s", opts.Name, errorString(resp, err))
	}

	server, resp, err := api.CreateLoadBalancerVirtualServer(ctx, loadbalancer.LbVirtualServer{
		Id:         opts.Name,
		Enabled:    true,
		IpAddress:  ip,
		IpProtocol: opts.Protocol,
		Ports:      opts.Ports,
		PoolId:     pool.Id,
	})

	if err != nil {
		return "", false, fmt.Errorf("unable to create virtual server %s: %s", opts.Name, errorString(resp, err))
	}

	lb := loadbalancer.LbService{
		DisplayName: opts.Name,
		Attachment: &common.ResourceReference{
			TargetType: "LogicalRouter",
			TargetId:   t1.Id,
		},
		Enabled:          true,
		ErrorLogLevel:    "INFO",
		Size:             "SMALL",
		VirtualServerIds: []string{server.Id},
	}

	_, resp, err = api.CreateLoadBalancerService(c.api.Context, lb)
	if err != nil {
		return "", false, fmt.Errorf("unable to create load balancer %s: %s", opts.Name, errorString(resp, err))
	}

	c.Infof("Created LoadBalancer service: %s/%s", server.Id, ip)
	return ip, false, nil
}

func (c *NSXClient) GetOrCreateHTTPHealthCheck(port string) (string, error) {
	id := fmt.Sprintf("http-%s", port)
	// ServicesApi.GetMonitors() is not implemented
	_, _, err := c.GET("/loadbalancer/monitors/" + id)

	if err == nil {
		// Get returned OK, so the monitor has been created already
		return id, nil
	}

	monitor, resp, err := c.api.ServicesApi.CreateLoadBalancerHttpMonitor(context.TODO(), loadbalancer.LbHttpMonitor{
		Id:                  id,
		DisplayName:         id,
		ResourceType:        "LbHttpMonitor",
		ResponseStatusCodes: []int32{200, 300, 301, 302, 304, 307, 404},
		MonitorPort:         port,
		RequestMethod:       "GET",
		RequestUrl:          "/",
		RequestVersion:      "HTTP_VERSION_1_1",
		Timeout:             15,
		Interval:            5,
		RiseCount:           5,
		FallCount:           3,
	})
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return monitor.Id, nil
}

func (c *NSXClient) GetOrCreateTCPHealthCheck(port string) (string, error) {
	id := fmt.Sprintf("tcp-%s", port)
	// ServicesApi.GetMonitors() is not implemented
	_, _, err := c.GET("/loadbalancer/monitors/" + id)

	if err == nil {
		// Get returned OK, so the monitor has been created already
		return id, nil
	}

	monitor, resp, err := c.api.ServicesApi.CreateLoadBalancerTcpMonitor(context.TODO(), loadbalancer.LbTcpMonitor{
		Id:           id,
		DisplayName:  id,
		ResourceType: "LbTcpMonitor",
		MonitorPort:  port,
		Timeout:      15,
		Interval:     5,
		RiseCount:    5,
		FallCount:    3,
	})
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return monitor.Id, nil
}
