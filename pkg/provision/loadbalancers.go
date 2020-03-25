package provision

import (
	"github.com/moshloop/platform-cli/pkg/nsx"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func provisionLoadbalancers(p *platform.Platform) (masters string, workers string, err error) {
	if p.NSX == nil || p.NSX.Disabled {
		return "", "", nil
	}

	nsxClient, err := p.GetNSXClient()
	if err != nil {
		return "", "", err
	}

	masters, err = nsxClient.CreateLoadBalancer(nsx.LoadBalancerOptions{
		Name:     p.Name + "-masters",
		IPPool:   p.NSX.LoadBalancerIPPool,
		Protocol: "TCP",
		Ports:    []string{"6443"},
		Tier0:    p.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": p.Name + "-masters",
		},
	})
	if err != nil {
		return "", "", err
	}
	workers, err = nsxClient.CreateLoadBalancer(nsx.LoadBalancerOptions{
		Name:     p.Name + "-workers",
		IPPool:   p.NSX.LoadBalancerIPPool,
		Protocol: "TCP",
		Ports:    []string{"80", "443"},
		Tier0:    p.NSX.Tier0,
		MemberTags: map[string]string{
			"Role": p.Name + "-workers",
		},
	})
	if err != nil {
		return "", "", err
	}
	return masters, workers, nil
}
