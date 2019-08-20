package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kapi "k8s.io/client-go/tools/clientcmd/api"
	"time"
)

type Platform struct {
	types.PlatformConfig
	ctx     context.Context
	session *vmware.Session
}

func (platform *Platform) GetVMs() (map[string]*VM, error) {
	var vms = make(map[string]*VM)
	list, err := platform.session.Finder.VirtualMachineList(
		platform.ctx, fmt.Sprintf("%s-%s-*", platform.HostPrefix, platform.Name))
	if err != nil {
		return nil, err
	}
	for _, vm := range list {
		item := &VM{
			Platform: platform,
			ctx:      platform.ctx,
			vm:       vm,
		}
		item.Name = vm.Name()
		vms[vm.Name()] = item
	}
	return vms, nil
}

func (platform *Platform) WaitFor() error {
	for {
		if len(platform.GetMasterIPs()) > 0 {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (platform *Platform) OpenViaEnv() error {
	platform.ctx = context.TODO()
	session, err := vmware.GetSessionFromEnv()
	if err != nil {
		return err
	}
	platform.session = session
	return nil
}

func (platform *Platform) GetMasterIPs() []string {
	url := fmt.Sprintf("http://%s/v1/health/service/%s", platform.Consul, platform.Name)
	log.Infof("Finding masters via consul: %s\n", url)
	response, _ := utils.GET(url)
	var consul api.Consul
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

func CreateKubeConfig(platform *Platform, endpoint string) ([]byte, error) {
	userName := "kubernetes-admin"
	contextName := fmt.Sprintf("%s@%s", userName, platform.Name)
	cert, err := platform.Certificates.CA.ToCert().CreateCertificate("system:masters", "system:masters")
	if err != nil {
		return nil, err
	}
	cfg := kapi.Config{
		Clusters: map[string]*kapi.Cluster{
			platform.Name: {
				Server:                   "https://" + endpoint + ":6443",
				CertificateAuthorityData: []byte(platform.Certificates.CA.X509),
			},
		},
		Contexts: map[string]*kapi.Context{
			contextName: {
				Cluster:  platform.Name,
				AuthInfo: userName,
			},
		},
		AuthInfos: map[string]*kapi.AuthInfo{
			userName: {
				ClientKeyData:         cert.EncodedPrivateKey(),
				ClientCertificateData: cert.EncodedCertificate(),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.Write(cfg)
}

func (platform *Platform) GetClientset() (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", platform.Name)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfigOrDie(cfg), nil
}
