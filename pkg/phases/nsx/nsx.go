package nsx

import (
	"fmt"

	"github.com/flanksource/commons/certs"

	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "nsx-system-operator"
	CertName  = "nsx-secret"
)

func Install(p *platform.Platform) error {
	if p.NSX == nil || p.NSX.Disabled || p.NSX.CNIDisabled {
		if err := p.DeleteSpecs(Namespace, "nsx-operator.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if !p.HasSecret(Namespace, CertName) {
		cert := certs.NewCertificateBuilder("kubernetes-client").Certificate
		cert, err := p.GetCA().SignCertificate(cert, 10)
		if err != nil {
			return fmt.Errorf("install: failed to sign certificate: %v", err)
		}

		if err := p.CreateOrUpdateSecret(CertName, Namespace, cert.AsTLSSecret()); err != nil {
			return fmt.Errorf("install: failed to create/update secret: %v", err)
		}
	}

	yes := true
	p.NSX.NsxV3.Insecure = &yes
	p.NSX.NsxCOE.Cluster = p.Name

	if err := p.ApplySpecs(Namespace, "nsx-operator.yaml"); err != nil {
		return fmt.Errorf("install: failed to apply specs: %v", err)
	}

	return nil
}
