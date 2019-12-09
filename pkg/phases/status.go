package phases

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/utils"
)

func Status(p *platform.Platform) error {

	if err := p.OpenViaEnv(); err != nil {
		return err
	}

	vmList, err := p.GetVMs()
	if err != nil {
		return err
	}

	vms := make(map[string]map[string]string)
	for _, vm := range vmList {
		attributes, err := vm.GetAttributes()
		if err != nil {
			attributes["error"] = fmt.Sprintf("%s", err)
		}
		ip, err := vm.GetIP(1 * time.Second)
		if err != nil {
			attributes["ip"] = fmt.Sprintf("Error: %v", err)
		} else {
			attributes["ip"] = ip
		}
		vms[vm.Name] = attributes
	}

	client, err := p.GetClientset()

	if err != nil {
		return nil
	}

	log.Infof("Listing nodes")
	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("%+v", err)
		return err
	}
	for _, node := range list.Items {
		attributes := vms[node.Name]
		delete(vms, node.Name)
		resources := fmt.Sprintf("%s (%s) cpu: %d/%d  mem: %s/%s", console.Greenf("%s", node.Name), attributes["ip"],
			node.Status.Allocatable.Cpu().Value(), node.Status.Capacity.Cpu().Value(),
			gb(node.Status.Allocatable.Memory().Value()),
			gb(node.Status.Capacity.Memory().Value()))
		fmt.Printf(resources + utils.Interpolate(
			"{{.Phase}} {{.NodeInfo.OperatingSystem}} {{.NodeInfo.OSImage}} {{.NodeInfo.KernelVersion}} {{.NodeInfo.ContainerRuntimeVersion}}\n",
			node.Status))
	}

	for vm, _ := range vms {
		fmt.Printf("%s VM not in cluster\n", console.Redf(vm))
	}
	return nil
}

func gb(bytes int64) string {
	return fmt.Sprintf("%.00d", bytes/1024/1024/1024)
}
