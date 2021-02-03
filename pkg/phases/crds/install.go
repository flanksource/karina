package crds

import (
	"sync"

	"github.com/flanksource/karina/pkg/platform"
)

func Install(p *platform.Platform) error {
	f := func() bool { return false }
	crds := map[string]func() bool{
		"cert-manager":      f,
		"service-monitor":   f,
		"canary-checker":    f,
		"template-operator": f,
		"git-operator":      f,
		"platform-operator": f,
		"prometheus":        f,
		"sealed-secrets":    f,
		"antrea":            func() bool { return p.Antrea == nil || p.Antrea.IsDisabled() },
		"calico":            func() bool { return p.Calico == nil || p.Calico.IsDisabled() },
		"helm-operator":     func() bool { return len(p.GitOps) == 0 },
		"eck":               p.ECK.IsDisabled,
		"gatekeeper":        p.Gatekeeper.IsDisabled,
		"postgres-db":       p.PostgresOperator.IsDisabled,
		"postgres-operator": p.PostgresOperator.IsDisabled,
		"rabbitmq":          p.RabbitmqOperator.IsDisabled,
		"tekton":            p.Tekton.IsDisabled,
		"velero":            p.Velero.IsDisabled,
		"vpa":               p.VPA.IsDisabled,
		"kiosk":             p.Kiosk.IsDisabled,
		"istio":             p.IstioOperator.IsDisabled,
	}

	wg := sync.WaitGroup{}
	for crd, fn := range crds {
		if fn() {
			continue
		}
		wg.Add(1)
		_crd := crd
		go func() {
			defer wg.Done()
			if err := p.ApplySpecs("", "crd/"+_crd+".yaml"); err != nil {
				p.Errorf("Error creating CRD: %s: %v", _crd, err)
			}
		}()
	}
	wg.Wait()
	return nil
}
