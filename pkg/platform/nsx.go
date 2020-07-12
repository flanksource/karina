package platform

import (
	"context"
	"fmt"

	nsxapi "github.com/flanksource/karina/pkg/nsx"
	"github.com/flanksource/karina/pkg/types"
)

type NSXProvider struct {
	nsxapi.NSXClient
}

func NewNSXProvider(platform *Platform) (*NSXProvider, error) {
	switch platform.MasterDiscovery.(type) {
	case *NSXProvider:
		return platform.MasterDiscovery.(*NSXProvider), nil
	}
	if platform.NSX == nil || platform.NSX.Disabled {
		return nil, fmt.Errorf("NSX not configured or disabled")
	}
	if platform.NSX.NsxV3 == nil || len(platform.NSX.NsxV3.NsxAPIManagers) == 0 {
		return nil, fmt.Errorf("nsx_v3.nsx_api_managers not configured")
	}

	client := &nsxapi.NSXClient{
		Logger:   platform.Logger,
		Host:     platform.NSX.NsxV3.NsxAPIManagers[0],
		Username: platform.NSX.NsxV3.NsxAPIUser,
		Password: platform.NSX.NsxV3.NsxAPIPass,
	}
	platform.Debugf("Connecting to NSX-T %s@%s", client.Username, client.Host)

	if err := client.Init(); err != nil {
		return nil, fmt.Errorf("getNSXClient: failed to init client: %v", err)
	}
	nsx := &NSXProvider{}
	nsx.NSXClient = *client
	return nsx, nil
}

func (nsx *NSXProvider) String() string {
	return fmt.Sprintf("NSX[%s]", nsx.NSXClient.Host)
}

func (nsx *NSXProvider) BeforeProvision(platform *Platform, machine *types.VM) error {
	return nil
}

func (nsx *NSXProvider) AfterProvision(platform *Platform, vm types.Machine) error {
	if platform.NSX == nil || platform.NSX.Disabled {
		return nil
	}
	ctx := context.TODO()

	ports, err := nsx.GetLogicalPorts(ctx, vm.Name())
	if err != nil {
		go vm.Terminate()
		return fmt.Errorf("failed to find ports for %s: %v", vm.Name(), err)
	}
	if len(ports) != 2 {
		go vm.Terminate()
		return fmt.Errorf("expected to find 2 ports, found %d \n%+v", len(ports), ports)
	}
	managementNic := make(map[string]string)
	transportNic := make(map[string]string)

	for k, v := range vm.GetTags() {
		managementNic[k] = v
	}

	transportNic["ncp/node_name"] = vm.Name()
	transportNic["ncp/cluster"] = platform.Name

	if err := nsx.TagLogicalPort(ctx, ports[0].Id, managementNic); err != nil {
		go vm.Terminate()
		return fmt.Errorf("failed to tag management nic %s: %v", ports[0].Id, err)
	}
	if err := nsx.TagLogicalPort(ctx, ports[1].Id, transportNic); err != nil {
		go vm.Terminate()
		return fmt.Errorf("failed to tag transport nic %s: %v", ports[1].Id, err)
	}
	platform.Tracef("Tagged %s", vm)
	return nil
}

func (nsx *NSXProvider) GetExternalEndpoints(platform *Platform) ([]string, error) {
	endpoints := []string{}
	if platform.DNS.IsEnabled() {
		endpoints = append(endpoints, "k8s-api."+platform.Domain)
	}
	lb, err := nsx.GetLoadBalancer(platform.Name + "-masters")
	if err != nil {
		return nil, err
	}
	if lb != nil {
		endpoints = append(endpoints, lb.IpAddress)
	}
	platform.Tracef("Discovered %s masters via NSX", endpoints)
	return endpoints, nil
}

func (nsx *NSXProvider) GetControlPlaneEndpoint(platform *Platform) (string, error) {
	if platform.NSX == nil || platform.NSX.Disabled {
		return "", fmt.Errorf("NSX not configured")
	}

	masterDNS := fmt.Sprintf("k8s-api.%s", platform.Domain)
	masterIP, _, err := nsx.CreateLoadBalancer(nsxapi.LoadBalancerOptions{
		Name:     platform.Name + "-masters",
		IPPool:   platform.NSX.LoadBalancerIPPool,
		Protocol: nsxapi.TCPProtocol,
		Ports:    []string{"6443"},
		Tier0:    platform.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": platform.Name + "-masters",
		},
	})
	if err != nil {
		return "", err
	}
	if !platform.DNS.IsEnabled() {
		masterDNS = masterIP
	}

	if err := platform.GetDNSClient().Update(masterDNS, masterIP); err != nil {
		platform.Warnf("Failed to create DNS entry for %s, failing back to IP: %s: %v", masterDNS, masterIP, err)
		masterDNS = masterIP
	}

	workerDNS := fmt.Sprintf("*.%s", platform.Domain)
	workerIP, _, err := nsx.CreateLoadBalancer(nsxapi.LoadBalancerOptions{
		Name:     platform.Name + "-workers",
		IPPool:   platform.NSX.LoadBalancerIPPool,
		Protocol: nsxapi.TCPProtocol,
		Ports:    []string{"80", "443"},
		Tier0:    platform.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": platform.Name + "-workers",
		},
	})
	if err != nil {
		return "", err
	}

	if err := platform.GetDNSClient().Update(workerDNS, workerIP); err != nil {
		platform.Warnf("Failed to create DNS entry for %s: %v", workerDNS, err)
	}
	return masterDNS + ":6443", nil
}

func (nsx *NSXProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func (nsx NSXProvider) AfterTerminate(platform *Platform, machine types.Machine) error {
	return nil
}
