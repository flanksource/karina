package dns

import (
	"github.com/flanksource/commons/logger"
)

type DummyDNSClient struct {
	logger.Logger
	Zone string
}

func (d DummyDNSClient) Append(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)
	d.Debugf("[DNS Stub] Append %s.%s %v", domain, d.Zone, records)
	return nil
}
func (d DummyDNSClient) Get(domain string) ([]string, error) { return nil, nil }
func (d DummyDNSClient) Update(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)
	d.Debugf("[DNS Stub] Update %s.%s %v", domain, d.Zone, records)
	return nil
}
func (d DummyDNSClient) Delete(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)
	d.Debugf("[DNS Stub] Delete %s.%s %v", domain, d.Zone, records)
	return nil
}
