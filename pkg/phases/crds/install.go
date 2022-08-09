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
		"canary-checker":    p.CanaryChecker.IsDisabled,
		"template-operator": p.TemplateOperator.IsDisabled,
		"git-operator":      p.GitOperator.IsDisabled,
		"platform-operator": p.PlatformOperator.IsDisabled,
		"prometheus":        f,
		"sealed-secrets":    p.SealedSecrets.IsDisabled,
		"antrea":            p.Antrea.IsDisabled,
		"calico":            p.Calico.IsDisabled,
		"argocd-operator":   p.ArgocdOperator.IsDisabled,
		"argo-rollouts":     p.ArgoRollouts.IsDisabled,
		"eck":               p.ECK.IsDisabled,
		"grafana-operator":  p.Monitoring.IsDisabled,
		"mongo-db":          p.MongodbOperator.IsDisabled, // TODO: Make this depends on MongoDB Operator instead of Keptn once MongoDB Operator is implemented: https://github.com/flanksource/karina/issues/658
		"mongodb-operator":  p.MongodbOperator.IsDisabled,
		"gatekeeper":        p.Gatekeeper.IsDisabled,
		"postgresql-db":     p.PostgresOperator.IsDisabled,
		"postgres-operator": p.PostgresOperator.IsDisabled,
		"rabbitmq":          p.RabbitmqOperator.IsDisabled,
		"redis":             p.RedisOperator.IsDisabled,
		"redis-db":          p.RedisOperator.IsDisabled,
		"velero":            p.Velero.IsDisabled,
		"vpa":               p.VPA.IsDisabled,
		"istio":             p.IstioOperator.IsDisabled,
		"logs-exporter":     p.LogsExporter.IsDisabled,
		"karina-operator":   p.KarinaOperator.IsDisabled,
		"flux":              func() bool { return p.Flux != nil && !p.Flux.Enabled },
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
