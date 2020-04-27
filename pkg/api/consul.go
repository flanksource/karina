package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/net"
)

type Consul struct {
	logger.Logger
	Host, Service string
}

func (consul Consul) GetMembers() []string {
	url := fmt.Sprintf("http://%s/v1/health/service/%s", consul.Host, consul.Service)
	consul.Tracef("Finding masters via consul: %s\n", url)
	response, _ := net.GET(url)
	var resp consulResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		fmt.Println(err)
	}
	var addresses []string
node:
	for _, node := range resp {
		for _, check := range node.Checks {
			if check.Status != "passing" {
				consul.Tracef("skipping unhealthy node %s -> %s", node.Node.Address, check.Status)
				continue node
			}
		}
		addresses = append(addresses, node.Node.Address)
	}
	return addresses
}

func (consul Consul) RemoveMember(name string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s/v1/catalog/deregister", consul.Host), strings.NewReader(fmt.Sprintf("{\"Node\": \"%s\"}", name)))

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to remove consul member %s: %s", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to remove consul member %s: %s", name, resp.Status)
	}
	consul.Infof("Removed consul member %s: %s", name, resp.Status)
	return nil
}

type consulResponse []struct {
	Node struct {
		ID              string `json:"ID"`
		Node            string `json:"Node"`
		Address         string `json:"Address"`
		Datacenter      string `json:"Datacenter"`
		TaggedAddresses struct {
			Lan string `json:"lan"`
			Wan string `json:"wan"`
		} `json:"TaggedAddresses"`
		Meta struct {
			ConsulNetworkSegment string `json:"consul-network-segment"`
		} `json:"Meta"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"Node"`
	Service struct {
		ID      string        `json:"ID"`
		Service string        `json:"Service"`
		Tags    []interface{} `json:"Tags"`
		Address string        `json:"Address"`
		Meta    interface{}   `json:"Meta"`
		Port    int           `json:"Port"`
		Weights struct {
			Passing int `json:"Passing"`
			Warning int `json:"Warning"`
		} `json:"Weights"`
		EnableTagOverride bool   `json:"EnableTagOverride"`
		ProxyDestination  string `json:"ProxyDestination"`
		Proxy             struct {
		} `json:"Proxy"`
		Connect struct {
		} `json:"Connect"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"Service"`
	Checks []struct {
		Node        string        `json:"Node"`
		CheckID     string        `json:"CheckID"`
		Name        string        `json:"Name"`
		Status      string        `json:"Status"`
		Notes       string        `json:"Notes"`
		Output      string        `json:"Output"`
		ServiceID   string        `json:"ServiceID"`
		ServiceName string        `json:"ServiceName"`
		ServiceTags []interface{} `json:"ServiceTags"`
		Definition  struct {
		} `json:"Definition"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"Checks"`
}
