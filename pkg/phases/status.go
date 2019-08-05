package phases

import (
	"fmt"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Status(p *platform.Platform) error {

	vms, err := p.GetVMs()
	if err != nil {
		return err
	}
	for _, vm := range vms {
		ip, err := vm.WaitForIP()
		if err != nil {
			fmt.Printf("%s: %s\n", vm.Name, err)
		} else {
			fmt.Printf("%s: %s\n", vm.Name, ip)
		}
	}

	client, err := p.GetClientset()

	if err != nil {
		return nil
	}

	list, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("%+v", err)
		return err
	}
	for _, node := range list.Items {

		resources := fmt.Sprintf("%s cpu: %d/%d  mem: %s/%s", node.Name,
			node.Status.Allocatable.Cpu().Value(), node.Status.Capacity.Cpu().Value(),
			gb(node.Status.Allocatable.Memory().Value()),
			gb(node.Status.Capacity.Memory().Value()))
		fmt.Printf(resources + utils.Interpolate(
			"{{.Phase}} {{.NodeInfo.OperatingSystem}} {{.NodeInfo.OSImage}} {{.NodeInfo.KernelVersion}} {{.NodeInfo.ContainerRuntimeVersion}}\n",
			node.Status))
	}

	return nil
}

func gb(bytes int64) string {
	return fmt.Sprintf("%.00d", bytes/1024/1024/1024)
}
