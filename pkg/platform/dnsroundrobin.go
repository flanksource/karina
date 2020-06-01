package platform

import (
	"fmt"

	"github.com/flanksource/karina/pkg/client/dns"
	"github.com/flanksource/karina/pkg/types"
)

type DnsProvider struct {
	dns.Client
}

func NewDnsProvider(client dns.Client) DnsProvider {
	provider := DnsProvider{}
	provider.Client = client
	return provider
}

func (dns DnsProvider) BeforeProvision(platform *Platform, machine types.VM) error { return nil }
func (dns DnsProvider) AfterProvision(platform *Platform, machine types.Machine) error {
	zone := "*."
	if machine.GetTags()["Role"] == platform.Name+"-masters" {
		zone = "k8s-api."
	}

	if err := dns.Append(zone+platform.Domain, machine.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", machine.IP())
	}

	return nil
}

func (dns DnsProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	zone := "*."
	if machine.GetTags()["Role"] == platform.Name+"-masters" {
		zone = "k8s-api."
	}

	if err := dns.Delete(zone+platform.Domain, machine.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", machine.IP())
	}
	return nil
}

func (dns DnsProvider) AfterTerminate(platform *Platform, machine types.Machine) error { return nil }

func (dns DnsProvider) GetControlPlaneEndpoint(platform *Platform) (string, error) {
	return fmt.Sprintf("k8s-api.%s", platform.Domain), nil
}
