package dns

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// Map of supported TSIG algorithms
var tsigAlgs = map[string]string{
	"hmac-md5":    dns.HmacMD5,
	"hmac-sha1":   dns.HmacSHA1,
	"hmac-sha256": dns.HmacSHA256,
	"hmac-sha512": dns.HmacSHA512,
}

type DNSClient struct {
	KeyName    string
	Nameserver string
	Key        string
	Algorithm  string
	Insecure   bool
}

func (client DNSClient) Append(zone, domain string, records ...string) error {

	m := new(dns.Msg)
	m.SetUpdate(zone)

	if !dns.IsFqdn(zone) {
		return fmt.Errorf("zone is not a fqdn")
	}

	for _, record := range records {
		RR := fmt.Sprintf("%s %d %s %s", domain, 60, "IN A", record)
		log.Infof("Adding RR: %s to %s", RR, zone)

		rr, err := dns.NewRR(RR)
		if err != nil {
			return fmt.Errorf("failed to build RR: %v", err)
		}

		m.Insert([]dns.RR{rr})
	}

	return client.sendMessage(zone, m)
}

func (client DNSClient) Get(domain string) ([]string, error) {

	m := new(dns.Msg)
	m.SetAxfr(domain)
	if client.Insecure {
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

func (client DNSClient) sendMessage(zone string, msg *dns.Msg) error {

	log.Debugf("KeyName=%s, Key=%s, Algorithm=%s\n Zone=%s, Nameserver=%s", client.KeyName, client.Key, client.Algorithm, msg.Question, client.Nameserver)

	c := new(dns.Client)
	c.SingleInflight = true

	if !client.Insecure {
		c.TsigSecret = map[string]string{zone: client.Key}
		msg.SetTsig(zone, tsigAlgs[client.Algorithm], 7500, time.Now().Unix())
	}

	resp, _, err := c.Exchange(msg, client.Nameserver)
	if err != nil {
		return fmt.Errorf("error in dns.Client.Exchange: %s", err)
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("bad return code: %s", dns.RcodeToString[resp.Rcode])
	}

	return nil
}

func (client DNSClient) Update(domain string, records ...string) error {
	return nil
}
