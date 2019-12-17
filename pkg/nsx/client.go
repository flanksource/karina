package nsx

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
)

type NSXClient struct {
	api                      *nsxt.APIClient
	Username, Password, Host string
	RemoteAuth               bool
}

func (c *NSXClient) Init() error {
	if c.api != nil {
		return nil
	}

	retriesConfig := nsxt.ClientRetriesConfiguration{
		MaxRetries:      3,
		RetryMinDelay:   1,
		RetryMaxDelay:   5,
		RetryOnStatuses: []int{429, 503},
	}

	cfg := nsxt.Configuration{

		UserName:   c.Username,
		Password:   c.Password,
		BasePath:   "/api/v1",
		Host:       c.Host,
		Scheme:     "https",
		UserAgent:  "platform-cli",
		RemoteAuth: c.RemoteAuth,
		DefaultHeader: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)),
		},
		Insecure:             true,
		RetriesConfiguration: retriesConfig,
	}

	client, err := nsxt.NewAPIClient(&cfg)
	if err != nil {
		return err
	}
	c.api = client
	return nil
}

func (c *NSXClient) Ping() (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("need to called .Init() first")
	}
	props, resp, err := c.api.NsxComponentAdministrationApi.ReadNodeProperties(c.api.Context)
	if err != nil {
		return "", fmt.Errorf("Error pinging: %v: resp: %v, req: %v", err, resp.Header, resp.Request.Header)
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Failed to get version: %s", resp.Status)
	}
	return props.NodeVersion, nil
}

func errorString(resp *http.Response, err error) string {
	s := fmt.Sprintf("%s", err)
	if id, ok := resp.Header["X-Nsx-Requestid"]; ok {
		s += fmt.Sprintf("%s reqid=%s", s, id)
	}
	return s

}
func (c *NSXClient) TagNics(ctx context.Context, id string, tags map[string]string) error {
	log.Debugf("Tagging %s: %v", id, tags)
	vms, resp, err := c.api.FabricApi.ListVirtualMachines(c.api.Context, map[string]interface{}{"displayName": id})
	if err != nil {
		return fmt.Errorf("Cannot get vifs for %s: %v", id, errorString(resp, err))
	}
	log.Tracef("Found %d vms\n", vms.ResultCount)
	for _, vm := range vms.Results {
		fmt.Printf("vm: %v\n", vm)
	}

	vifs, resp, err := c.api.FabricApi.ListVifs(c.api.Context, map[string]interface{}{"ownerVmId": vms.Results[0].ExternalId})
	if err != nil {
		return fmt.Errorf("Cannot get vifs for %s: %v", id, errorString(resp, err))
	}
	log.Tracef("Found %d vifs\n", vifs.ResultCount)
	for _, vif := range vifs.Results {
		ports, resp, err := c.api.LogicalSwitchingApi.ListLogicalPorts(c.api.Context, map[string]interface{}{"attachmentId": vif.LportAttachmentId})

		if err != nil {
			return fmt.Errorf("Unable to get port %s: %s", vif.LportAttachmentId, errorString(resp, err))
		}

		for _, port := range ports.Results {
			for k, v := range tags {
				port.Tags = append(port.Tags, common.Tag{
					Scope: k,
					Tag:   v,
				})
			}
			_, resp, err = c.api.LogicalSwitchingApi.UpdateLogicalPort(context.TODO(), port.Id, port)
			if err != nil {
				return fmt.Errorf("Unable to update port %s: %s", port.Id, errorString(resp, err))
			}
		}
	}
	return nil

}
