package nsx

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/flanksource/commons/logger"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	TcpProtocol = "TCP"
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

func (c *NSXClient) GET(path string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	url := fmt.Sprintf("%s://%s%s%s", c.cfg.Scheme, c.cfg.Host, c.cfg.BasePath, path)
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range c.cfg.DefaultHeader {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to list virtual servers: %s", errorString(resp, err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GET: Failed to read response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return body, fmt.Errorf(resp.Status)
	}
	return body, nil
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

// CreateLoadBalancer creates a new loadbalancer or returns the existing loadbalancer's IP
func (c *NSXClient) CreateLoadBalancer(opts LoadBalancerOptions) (string, bool, error) {
	ctx := c.api.Context
	api := c.api.ServicesApi
	routing := c.api.LogicalRoutingAndServicesApi

	// ServicesApi.GetVirtualServers() is not implemented
	body, err := c.GET("/loadbalancer/virtual-servers")
	if err != nil {
		return "", false, fmt.Errorf("failed to list existing virtual servers: %v", err)
	}

	var virtualServers virtualServersList

	if err := json.Unmarshal(body, &virtualServers); err != nil {
		return "", false, fmt.Errorf("failed to unmarshall existing virtual server list: %v", err)
	}

	for _, server := range virtualServers.Results {
		if server.DisplayName == opts.Name {
			c.Debugf("LoadBalancer %s found, returning its IP %s ", opts.Name, server.IpAddress)
			return server.IpAddress, true, nil
		}
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
	if opts.Protocol == TcpProtocol {
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
	_, err := c.GET("/loadbalancer/monitors/" + id)

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
	_, err := c.GET("/loadbalancer/monitors/" + id)

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
