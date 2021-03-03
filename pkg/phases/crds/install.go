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
		"helm-operator":     f,
		"antrea":            p.Antrea.IsDisabled,
		"calico":            p.Calico.IsDisabled,
		"argocd-operator":   p.ArgocdOperator.IsDisabled,
		"argo-rollouts":     p.ArgoRollouts.IsDisabled,
		"eck":               p.ECK.IsDisabled,
		"mongo-db":          p.Keptn.IsDisabled, // TODO: Make this depends on MongoDB Operator instead of Keptn once MongoDB Operator is implemented: https://github.com/flanksource/karina/issues/658
		"gatekeeper":        p.Gatekeeper.IsDisabled,
		"postgres-db":       p.PostgresOperator.IsDisabled,
		"postgres-operator": p.PostgresOperator.IsDisabled,
		"rabbitmq":          p.RabbitmqOperator.IsDisabled,
		"tekton":            p.Tekton.IsDisabled,
		"velero":            p.Velero.IsDisabled,
		"vpa":               p.VPA.IsDisabled,
		"kiosk":             p.Kiosk.IsDisabled,
		"istio":             p.IstioOperator.IsDisabled,
		"logs-exporter":     p.LogsExporter.IsDisabled,
		"karina-operator":   p.KarinaOperator.IsDisabled,
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
