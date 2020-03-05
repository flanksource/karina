package dns

import (
	log "github.com/sirupsen/logrus"
)

type DummyDNSClient struct {
	Zone string
}

func (d DummyDNSClient) Append(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)
	log.Debugf("[DNS Stub] Append %s.%s %v", domain, d.Zone, records)
	return nil
}
func (d DummyDNSClient) Get(domain string) ([]string, error) { return nil, nil }
func (d DummyDNSClient) Update(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)

	log.Debugf("[DNS Stub] Update %s.%s %v", domain, d.Zone, records)
	return nil
}
func (d DummyDNSClient) Delete(domain string, records ...string) error {
	domain = subdomain(domain, d.Zone)
	log.Debugf("[DNS Stub] Delete %s.%s %v", domain, d.Zone, records)
	return nil
}
