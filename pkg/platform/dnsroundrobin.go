package platform

import (
	"fmt"

	"github.com/flanksource/karina/pkg/client/dns"
	"github.com/flanksource/karina/pkg/types"
)

type DNSProvider struct {
	dns.Client
}

func NewDNSProvider(client dns.Client) DNSProvider {
	provider := DNSProvider{}
	provider.Client = client
	return provider
}

func (dns DNSProvider) String() string {
	return fmt.Sprintf("DNS(%s)", dns.Client)
}

func (dns DNSProvider) BeforeProvision(platform *Platform, machine *types.VM) error { return nil }
func (dns DNSProvider) AfterProvision(platform *Platform, machine types.Machine) error {
	zone := "*."
	if platform.IsMaster(machine) {
		zone = "k8s-api."
	}

	if err := dns.Append(zone+platform.Domain, machine.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", machine.IP())
	}

	return nil
}

func (dns DNSProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	zone := "*."
	if platform.IsMaster(machine) {
		zone = "k8s-api."
	}

	if err := dns.Delete(zone+platform.Domain, machine.IP()); err != nil {
		platform.Warnf("Failed to update DNS for %s", machine.IP())
	}
	return nil
}

func (dns DNSProvider) AfterTerminate(platform *Platform, machine types.Machine) error { return nil }

func (dns DNSProvider) GetControlPlaneEndpoint(platform *Platform) (string, error) {
	return fmt.Sprintf("k8s-api.%s", platform.Domain), nil
}

func (dns DNSProvider) GetExternalEndpoints(platform *Platform) ([]string, error) {
	platform.Tracef("Using DNS endpoint for master discovery")
	return []string{"k8s-api." + platform.Domain}, nil
}
