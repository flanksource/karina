package dns

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/flanksource/commons/logger"
	"github.com/thoas/go-funk"
)

type Route53Client struct {
	logger.Logger
	HostedZoneID         string
	AccessKey, SecretKey string
	Domain               string
	session              *session.Session
	svc                  *route53.Route53
}

func (r53 *Route53Client) Init() {
	if r53.session == nil {
		r53.session, _ = session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(r53.AccessKey, r53.SecretKey, ""),
		})
		r53.svc = route53.New(r53.session)
	}
}

func getResourceRecords(records ...string) []*route53.ResourceRecord {
	out := []*route53.ResourceRecord{}
	for _, record := range records {
		value := record
		out = append(out, &route53.ResourceRecord{
			Value: &value,
		})
	}
	return out
}

func normalize(name string) string {
	return strings.Replace(name, "\\052", "*", 1)
}

func (r53 *Route53Client) Append(domain string, records ...string) error {
	existing, err := r53.Get(domain)
	if err != nil {
		return fmt.Errorf("error getting existing records for domain %s, %v", domain, err)
	}
	return r53.Update(domain, append(existing, records...)...)
}

func (r53 *Route53Client) Get(domain string) ([]string, error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	a := "A"
	output, err := r53.svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		StartRecordType: &a,
		HostedZoneId:    aws.String(r53.HostedZoneID),
		StartRecordName: aws.String(domain),
	})
	if err != nil {
		return []string{}, fmt.Errorf("error getting records for %s: %v", domain, err)
	}

	var records []string
	for _, set := range output.ResourceRecordSets {
		if normalize(*set.Name) != domain || *set.Type != a {
			continue
		}
		for _, record := range set.ResourceRecords {
			records = append(records, *record.Value)
		}
	}

	r53.Tracef("lookup %s => %v", domain, records)
	return records, nil
}

func (r53 *Route53Client) Update(domain string, records ...string) error {
	rr := getResourceRecords(records...)
	ttl := int64(60)
	r53.Tracef("Updating %s domain to %v", domain, records)
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						ResourceRecords: rr,
						Name:            aws.String(domain),
						Type:            aws.String("A"),
						TTL:             &ttl,
					},
				},
			},
		},
		HostedZoneId: aws.String(r53.HostedZoneID),
	}

	_, err := r53.svc.ChangeResourceRecordSets(input)
	return err
}

func (r53 *Route53Client) Delete(domain string, records ...string) error {
	existing, err := r53.Get(domain)
	if err != nil {
		return fmt.Errorf("error getting existing records for domain %s, %v", domain, err)
	}
	return r53.Update(domain, funk.Subtract(existing, records).([]string)...)
}
