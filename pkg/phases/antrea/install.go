package antrea

import (
	"github.com/flanksource/karina/pkg/platform"
)

const Namespace = "kube-system"

func Install(p *platform.Platform) error {
	if (p.Calico != nil && !p.Calico.IsDisabled()) && (p.Antrea != nil && !p.Antrea.IsDisabled()) {
		p.Fatalf("both calico and antrea are enabled. Please disable one of them to continue")
	}

	if p.Antrea == nil || p.Antrea.IsDisabled() {
		if err := p.DeleteSpecs(Namespace, "antrea.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	secret := p.GetSecret("kube-system", "antrea-controller-tls")
	if secret != nil {
		data := *secret
		caBytes := data["ca.crt"]
		crtBytes := data["tls.crt"]
		keyBytes := data["tls.key"]
		if len(caBytes) > 0 && len(crtBytes) > 0 && len(keyBytes) > 0 {
			p.Antrea.IsCertReady = true
		}
	}

	return p.ApplySpecs(Namespace, "antrea.yaml")
}
