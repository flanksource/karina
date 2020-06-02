package platform

import (
	"errors"
	"fmt"

	"github.com/flanksource/karina/pkg/api"
	"github.com/flanksource/karina/pkg/types"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
)

type ConsulProvider struct {
	api.Consul
}

func NewConsulProvider(platform *Platform) ConsulProvider {
	provider := ConsulProvider{}
	provider.Consul = api.Consul{
		Logger:  platform.Logger,
		Host:    fmt.Sprintf("http://%s:8500", platform.Consul),
		Service: platform.Name,
	}
	return provider
}

func (consul ConsulProvider) BeforeProvision(platform *Platform, machine *types.VM) error {
	if platform.Datacenter == "" {
		return errors.New("Must specify a platform datacenter")
	}
	if platform.Consul == "" {
		return errors.New("Must specify a consul host")
	}
	if platform.IsMaster(*machine) {
		createConsulService(machine.Name, platform, machine.Konfigadm)
	}

	createClientSideLoadbalancers(platform, machine.Konfigadm)
	return nil
}

func (consul ConsulProvider) AfterProvision(platform *Platform, machine types.Machine) error {
	return nil
}

func (consul ConsulProvider) BeforeTerminate(platform *Platform, machine types.Machine) error {
	return consul.RemoveMember(machine.Name())
}

func (consul ConsulProvider) AfterTerminate(platform *Platform, machine types.Machine) error {
	return nil
}

func (consul ConsulProvider) GetControlPlaneEndpoint(platform *Platform) (string, error) {
	return "localhost:8443", nil
}

func (consul ConsulProvider) GetExternalEndpoints(platform *Platform) ([]string, error) {
	members := consul.GetMembers()
	platform.Tracef("Discovered %s masters via consul", members)
	return members, nil

}

func (consul ConsulProvider) String() string {
	return fmt.Sprintf("Consul(%s)", consul.Host)
}

// createConsulService derives the initial consul config for a cluster from its platform
// config and adds it to its konfigadm files
func createConsulService(hostname string, platform *Platform, cfg *konfigadm.Config) {
	cfg.Files["/etc/kubernetes/consul/api.json"] = fmt.Sprintf(`
{
	"leave_on_terminate": true,
  "rejoin_after_leave": true,
	"service": {
		"id": "%s",
		"name": "%s",
		"address": "",
		"check": {
			"id": "api-server",
			"name": " TCP on port 6443",
			"tcp": "localhost:6443",
			"interval": "120s",
			"timeout": "60s"
		},
		"port": 6443,
		"enable_tag_override": false
	}
}
	`, hostname, platform.Name)
}

// createClientSideLoadbalancers derives the client side loadbalancer configs for a cluster from its platform
// config and adds it to its konfigadm containers
func createClientSideLoadbalancers(platform *Platform, cfg *konfigadm.Config) {
	cfg.Containers = append(cfg.Containers, konfigadm.Container{
		Image: platform.GetImagePath("docker.io/consul:1.3.1"),
		Env: map[string]string{
			"CONSUL_CLIENT_INTERFACE": "ens160",
			"CONSUL_BIND_INTERFACE":   "ens160",
		},
		Args:       fmt.Sprintf("agent -join=%s:8301 -datacenter=%s -data-dir=/consul/data -domain=consul -config-dir=/consul-configs", platform.Consul, platform.Datacenter),
		DockerOpts: "--net host",
		Volumes: []string{
			"/etc/kubernetes/consul:/consul-configs",
		},
	}, konfigadm.Container{
		Image:      platform.GetImagePath("docker.io/moshloop/tcp-loadbalancer:0.1"),
		Service:    "haproxy",
		DockerOpts: "--net host -p 8443:8443",
		Env: map[string]string{
			"CONSUL_CONNECT": platform.Consul + ":8500",
			"SERVICE_NAME":   platform.Name,
			"PORT":           "8443",
		},
	})
}
