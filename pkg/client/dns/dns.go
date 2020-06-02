package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/miekg/dns"
)

var tsigAlgs = map[string]string{
	"hmac-md5":    dns.HmacMD5,
	"hmac-sha1":   dns.HmacSHA1,
	"hmac-sha256": dns.HmacSHA256,
	"hmac-sha512": dns.HmacSHA512,
}

type Client interface {
	Append(domain string, records ...string) error
	Get(domain string) ([]string, error)
	Update(domain string, records ...string) error
	Delete(domain string, records ...string) error
}

type DynamicDNSClient struct {
	logger.Logger
	KeyName    string
	Zone       string
	Nameserver string
	Key        string
	Algorithm  string
	Insecure   bool
}

func (client DynamicDNSClient) String() string {
	return fmt.Sprintf("DNS(%s@%s)", client.Zone, client.Nameserver)
}

func (client DynamicDNSClient) Append(domain string, records ...string) error {
	domain = subdomain(domain, client.Zone)
	client.Debugf("Appending %s.%s %s", domain, client.Zone, records)
	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	for _, record := range records {
		rr, err := newRR(domain, client.Zone, 60, "A", record)
		if err != nil {
			return fmt.Errorf("append: failed to get new RR: %v", err)
		}
		m.Insert([]dns.RR{*rr})
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
			}
			return nil, fmt.Errorf("AXFR error: %v", e.Error)
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
	client.Debugf("Updating %s.%s %s", domain, client.Zone, records)
	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	rr, err := newRR(domain, client.Zone, 0, "ANY", "")
	if err != nil {
		return fmt.Errorf("update: failed to get new RR: %v", err)
	}
	m.RemoveRRset([]dns.RR{*rr})

	for _, record := range records {
		rr, err := newRR(domain, client.Zone, 60, "A", record)
		if err != nil {
			return fmt.Errorf("update: failed to get new RR: %v", err)
		}
		m.Insert([]dns.RR{*rr})
	}
	return client.sendMessage(client.Zone, m)
}

func (client DynamicDNSClient) Delete(domain string, records ...string) error {
	domain = subdomain(domain, client.Zone)
	client.Debugf("Removing %s.%s %s", domain, client.Zone, records)

	m := new(dns.Msg)
	m.SetUpdate(client.Zone + ".")

	for _, record := range records {
		if record == "*" {
			rr, err := newRR(domain, client.Zone, 0, "ANY", "")
			if err != nil {
				return fmt.Errorf("delete: failed to get new RR: %v", err)
			}
			m.RemoveRRset([]dns.RR{*rr})
		} else {
			rr, err := newRR(domain, client.Zone, 0, "A", record)
			if err != nil {
				return fmt.Errorf("delete: failed to get new RR: %v", err)
			}
			m.Remove([]dns.RR{*rr})
		}
	}
	return client.sendMessage(client.Zone, m)
}

// nolint: unparam
func (client DynamicDNSClient) sendMessage(zone string, msg *dns.Msg) error {
	c := new(dns.Client)
	c.SingleInflight = true
	c.Net = "tcp"

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
	rr, err := dns.NewRR(RR)
	if err != nil {
		return nil, fmt.Errorf("newRR failed: %v", err)
	}
	return &rr, nil
}

func subdomain(domain, zone string) string {
	return strings.ReplaceAll(domain, "."+zone, "")
}
