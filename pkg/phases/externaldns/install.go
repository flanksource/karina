package externaldns

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "platform-system"
)

func Install(p *platform.Platform) error {
	if p.ExternalDNS.IsDisabled() {
		return p.DeleteSpecs(Namespace, "external-dns.yaml")
	}

	return p.ApplySpecs(Namespace, "external-dns.yaml")
}
