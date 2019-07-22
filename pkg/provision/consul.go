package provision

import (
	"encoding/json"
	"fmt"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"time"
	// "github.com/moshloop/konfigadm/pkg/utils"
	"github.com/moshloop/platform-cli/pkg/types"
)

func WaitForIP(platform types.PlatformConfig, ip string) error {
	for {
		for _, healthy := range GetMasterIPs(platform) {
			if healthy == ip {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func GetMasterIPs(platform types.PlatformConfig) []string {
	url := fmt.Sprintf("http://%s/v1/health/service/%s", platform.Consul, platform.Name)
	log.Infof("Finding masters via consul: %s\n", url)
	response, _ := utils.GET(url)
	var consul Consul
	if err := json.Unmarshal(response, &consul); err != nil {
		fmt.Println(err)
	}
	var addresses []string
node:
	for _, node := range consul {
		for _, check := range node.Checks {
			if check.Status != "passing" {
				log.Tracef("skipping unhealthy node %s -> %s", node.Node.Address, check.Status)
				continue node
			}
		}
		addresses = append(addresses, node.Node.Address)
	}
	return addresses
}

type Consul []struct {
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
