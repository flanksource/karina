package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var tsigAlgs = map[string]string{
	"hmac-md5":    dns.HmacMD5,
	"hmac-sha1":   dns.HmacSHA1,
	"hmac-sha256": dns.HmacSHA256,
	"hmac-sha512": dns.HmacSHA512,
}

type DNSClient interface {
	Append(domain string, records ...string) error
	Get(domain string) ([]string, error)
	Update(domain string, records ...string) error
	Delete(domain string, records ...string) error
}

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

type DynamicDNSClient struct {
	KeyName    string
	Zone       string
	Nameserver string
	Key        string
	Algorithm  string
	Insecure   bool
}

func (client DynamicDNSClient) Append(domain string, records ...string) error {
	domain = subdomain(domain, client.Zone)
	log.Debugf("Appending %s.%s %s", domain, client.Zone, records)
	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	for _, record := range records {
		if rr, err := newRR(domain, client.Zone, 60, "A", record); err != nil {
			return err
		} else {
			m.Insert([]dns.RR{*rr})
		}
	}
	return client.sendMessage(client.Zone, m)
}

func (client DynamicDNSClient) Get(domain string) ([]string, error) {

	m := new(dns.Msg)
	m.SetAxfr(domain)
	if !client.Insecure {
		m.SetTsig(client.KeyName, tsigAlgs[client.Algorithm], 300, time.Now().Unix())
	}

	t := new(dns.Transfer)
	if !client.Insecure {
		t.TsigSecret = map[string]string{client.KeyName: client.Key}
	}

	env, err := t.In(m, client.Nameserver)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch records via AXFR: %v", err)
	}

	var records []string
	for e := range env {
		if e.Error != nil {
			if e.Error == dns.ErrSoa {
				return nil, fmt.Errorf("AXFR error: unexpected response received from the server")
			} else {
				return nil, fmt.Errorf("AXFR error: %v", e.Error)
			}
		}
		for _, rr := range e.RR {
			switch rr.Header().Rrtype {
			case dns.TypeCNAME:
				records = append(records, rr.(*dns.CNAME).Target)

			case dns.TypeA:
				records = append(records, rr.(*dns.A).A.String())

			case dns.TypeAAAA:
				records = append(records, rr.(*dns.AAAA).AAAA.String())

			case dns.TypeTXT:
				records = append(records, rr.(*dns.TXT).String())
			default:
				continue // Unhandled record type
			}
		}
	}

	return records, nil
}

func (client DynamicDNSClient) Update(domain string, records ...string) error {
	domain = subdomain(domain, client.Zone)
	log.Debugf("Updating %s.%s %s", domain, client.Zone, records)
	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	if rr, err := newRR(domain, client.Zone, 0, "ANY", ""); err != nil {
		return err
	} else {
		m.RemoveRRset([]dns.RR{*rr})
	}

	for _, record := range records {
		if rr, err := newRR(domain, client.Zone, 60, "A", record); err != nil {
			return err
		} else {
			m.Insert([]dns.RR{*rr})
		}
	}
	return client.sendMessage(client.Zone, m)
}

func (client DynamicDNSClient) Delete(domain string, records ...string) error {
	domain = subdomain(domain, client.Zone)
	log.Debugf("Removing %s.%s %s", domain, client.Zone, records)

	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	for _, record := range records {
		if record == "*" {
			if rr, err := newRR(domain, client.Zone, 0, "ANY", ""); err != nil {
				return err
			} else {
				m.RemoveRRset([]dns.RR{*rr})
			}
		} else {
			if rr, err := newRR(domain, client.Zone, 0, "A", record); err != nil {
				return err
			} else {
				m.Remove([]dns.RR{*rr})
			}
		}
	}
	return client.sendMessage(client.Zone, m)
}

func (client DynamicDNSClient) sendMessage(zone string, msg *dns.Msg) error {
	c := new(dns.Client)
	c.SingleInflight = true

	c.TsigSecret = map[string]string{client.KeyName: client.Key}
	msg.SetTsig(client.KeyName, tsigAlgs[client.Algorithm], 300, time.Now().Unix())

	resp, _, err := c.Exchange(msg, client.Nameserver)
	if err != nil {
		return fmt.Errorf("error in dns.Client.Exchange: %s", err)
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("bad return code: %s", dns.RcodeToString[resp.Rcode])
	}

	return nil
}

func newRR(domain string, zone string, ttl int, resourceType string, record string) (*dns.RR, error) {
	RR := strings.Trim(fmt.Sprintf("%s.%s %d %s %s", domain, zone, ttl, resourceType, record), " ")
	log.Tracef(RR)
	rr, err := dns.NewRR(RR)
	if err != nil {
		return nil, err
	}
	return &rr, nil
}

func subdomain(domain, zone string) string {
	return strings.ReplaceAll(domain, "."+zone, "")
}
