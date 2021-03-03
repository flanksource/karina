package vsphere

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	if platform.Vsphere == nil {
		return
	}
	client, _ := platform.GetClientset()

	if platform.Vsphere.CPIVersion != "" {
		kommons.TestDaemonSet(client, Namespace, "vsphere-cloud-controller-manager", test)
	}
	if platform.Vsphere.CSIVersion != "" {
		kommons.TestDeploy(client, Namespace, "vsphere-csi-controller", test)
	}
}
