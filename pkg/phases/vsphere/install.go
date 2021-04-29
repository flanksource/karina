package vsphere

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(platform *platform.Platform) error {
	if platform.Vsphere == nil {
		return nil
	}
	if platform.Vsphere.IsDisabled() {
		return nil
	}
	v := platform.Vsphere
	if err := platform.CreateOrUpdateSecret("vsphere-secrets", Namespace, platform.Vsphere.GetSecret()); err != nil {
		platform.Errorf("Failed to create vsphere secrets: %s", err)
	}
	if err := platform.CreateOrUpdateSecret("vsphere-config", Namespace, map[string][]byte{
		"vsphere.conf": []byte(fmt.Sprintf(`
[Global]
cluster-id = "%s"
port = "443"
insecure-flag = "true"
secret-name = "vsphere-secrets"
secret-namespace = "kube-system"

[VirtualCenter "%s"]
datacenters = "%s"
user = "%s"
password = "%s"
			`, platform.Name, v.Hostname, v.Datacenter, v.Username, v.Password)),
	}); err != nil {
		platform.Errorf("Failed to create vsphere config: %s", err)
	}

	if platform.Vsphere.CPIVersion != "" {
		if err := platform.ApplySpecs(Namespace, "vsphere-cpi.yaml"); err != nil {
			platform.Errorf("Failed to deploy vSphere CPI: %v", err)
		}
	}
	if platform.Vsphere.CSIVersion != "" {
		if err := platform.ApplySpecs(Namespace, "vsphere-csi.yaml"); err != nil {
			platform.Errorf("Failed to deploy vSphere CSI: %v", err)
		}
	}
	return nil
}
