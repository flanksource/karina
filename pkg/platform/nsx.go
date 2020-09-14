package platform

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	nsxapi "github.com/flanksource/karina/pkg/nsx"
	"github.com/flanksource/karina/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type NSXProvider struct {
	nsxapi.NSXClient
}

func terminate(platform *Platform, vm types.Machine) {
	platform.Errorf("terminating vm after failed NIC tagging %s", vm)
	_ = platform.DeleteNode(vm.Name())
	_ = vm.Terminate()
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

func (nsx *NSXProvider) tag(platform *Platform, vm types.Machine) error {
	ctx := context.TODO()

	ports, err := nsx.GetLogicalPorts(ctx, vm.Name())

	if err != nil {
		return fmt.Errorf("failed to find ports for %s: %v", vm.Name(), err)
	}
	if len(ports) != 2 {
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
		return fmt.Errorf("failed to tag management nic %s: %v", ports[0].Id, err)
	}
	if err := nsx.TagLogicalPort(ctx, ports[1].Id, transportNic); err != nil {
		return fmt.Errorf("failed to tag transport nic %s: %v", ports[1].Id, err)
	}
	platform.Tracef("Tagged %s", vm)
	return nil
}

func (nsx *NSXProvider) AfterProvision(platform *Platform, vm types.Machine) error {
	if platform.NSX == nil || platform.NSX.Disabled {
		return nil
	}

	err := backoff(func() error {
		return nsx.tag(platform, vm)
	}, platform.Logger, nil)

	if err != nil {
		go terminate(platform, vm)
		return err
	}
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

	masterDNS = updateDNS(platform, masterDNS, masterIP)

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

	updateDNS(platform, workerDNS, workerIP)
	return masterDNS + ":6443", nil
}

func updateDNS(platform *Platform, dns string, ip string) string {
	if !platform.DNS.IsEnabled() {
		return ip
	}
	lookupDNS := dns
	if strings.HasPrefix(dns, "*") {
		// Wildcard domains are set using "*", but a lookup for a * domain
		// will fail so we substitute it with a random domain
		lookupDNS = "random-wildcard" + dns[1:]
	}
	ips, err := platform.GetDNSClient().Get(lookupDNS)
	if err != nil {
		// try using the system resolver
		_ips, err := net.LookupIP(lookupDNS)
		if err == nil {
			for _, ip := range _ips {
				ips = append(ips, ip.To4().String())
			}
		} else {
			platform.Warnf("Failed lookup DNS entry for %s: %v", lookupDNS, err)
		}
	}
	if len(ips) == 0 {
		platform.Infof("Updating DNS %s: -> %s", dns, ip)
	} else if ips[0] != ip {
		platform.Infof("Updating DNS %s: from %s to %s", dns, ips[0], ip)
	}
	if len(ips) == 0 || ips[0] != ip {
		if err := platform.GetDNSClient().Update(dns, ip); err != nil {
			platform.Warnf("Failed to create DNS entry for %s, failing back to IP: %s: %v", dns, ip, err)
			return ip
		}
	}
	return dns
}

func (nsx *NSXProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func (nsx NSXProvider) AfterTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func backoff(fn func() error, log logger.Logger, backoffOpts *wait.Backoff) error {
	var returnErr *error
	if backoffOpts == nil {
		backoffOpts = &wait.Backoff{
			Duration: 500 * time.Millisecond,
			Factor:   2.0,
			Steps:    7,
		}
	}

	_ = wait.ExponentialBackoff(*backoffOpts, func() (bool, error) {
		err := fn()
		if err == nil {
			return true, nil
		}
		log.Warnf("retrying after error: %v", err)
		returnErr = &err
		return false, nil
	})
	if returnErr != nil {
		return *returnErr
	}
	return nil
}
