package order

import (
	"github.com/flanksource/karina/pkg/phases/antrea"
	"github.com/flanksource/karina/pkg/phases/argocdoperator"
	"github.com/flanksource/karina/pkg/phases/argorollouts"
	"github.com/flanksource/karina/pkg/phases/auditbeat"
	"github.com/flanksource/karina/pkg/phases/base"
	"github.com/flanksource/karina/pkg/phases/calico"
	"github.com/flanksource/karina/pkg/phases/canary"
	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/phases/configmapreloader"
	"github.com/flanksource/karina/pkg/phases/crds"
	"github.com/flanksource/karina/pkg/phases/dex"
	"github.com/flanksource/karina/pkg/phases/eck"
	"github.com/flanksource/karina/pkg/phases/elasticsearch"
	"github.com/flanksource/karina/pkg/phases/eventrouter"
	"github.com/flanksource/karina/pkg/phases/externaldns"
	"github.com/flanksource/karina/pkg/phases/filebeat"
	"github.com/flanksource/karina/pkg/phases/flux"
	"github.com/flanksource/karina/pkg/phases/gitoperator"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/phases/ingress"
	"github.com/flanksource/karina/pkg/phases/istiooperator"
	"github.com/flanksource/karina/pkg/phases/journalbeat"
	"github.com/flanksource/karina/pkg/phases/karinaoperator"
	"github.com/flanksource/karina/pkg/phases/kiosk"
	"github.com/flanksource/karina/pkg/phases/kpack"
	"github.com/flanksource/karina/pkg/phases/kuberesourcereport"
	"github.com/flanksource/karina/pkg/phases/kubewebview"
	"github.com/flanksource/karina/pkg/phases/logsexporter"
	"github.com/flanksource/karina/pkg/phases/minio"
	"github.com/flanksource/karina/pkg/phases/monitoring"
	"github.com/flanksource/karina/pkg/phases/nsx"
	"github.com/flanksource/karina/pkg/phases/opa"
	"github.com/flanksource/karina/pkg/phases/packetbeat"
	"github.com/flanksource/karina/pkg/phases/platformoperator"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/phases/pre"
	"github.com/flanksource/karina/pkg/phases/quack"
	"github.com/flanksource/karina/pkg/phases/rabbitmqoperator"
	"github.com/flanksource/karina/pkg/phases/redisoperator"
	"github.com/flanksource/karina/pkg/phases/registrycreds"
	"github.com/flanksource/karina/pkg/phases/s3uploadcleaner"
	"github.com/flanksource/karina/pkg/phases/sealedsecrets"
	"github.com/flanksource/karina/pkg/phases/stubs"
	"github.com/flanksource/karina/pkg/phases/tekton"
	"github.com/flanksource/karina/pkg/phases/templateoperator"
	"github.com/flanksource/karina/pkg/phases/vault"
	"github.com/flanksource/karina/pkg/phases/velero"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"
)

type DeployFn func(p *platform.Platform) error

var Phases = map[string]DeployFn{
	"pre":                  pre.Install,
	"argocd-operator":      argocdoperator.Deploy,
	"argo-rollouts":        argorollouts.Deploy,
	"antrea":               antrea.Install,
	"auditbeat":            auditbeat.Deploy,
	"base":                 base.Install,
	"calico":               calico.Install,
	"canary":               canary.Deploy,
	"configmap-reloader":   configmapreloader.Deploy,
	"crds":                 crds.Install,
	"dex":                  dex.Install,
	"eck":                  eck.Deploy,
	"elasticsearch":        elasticsearch.Deploy,
	"eventrouter":          eventrouter.Deploy,
	"externaldns":          externaldns.Install,
	"filebeat":             filebeat.Deploy,
	"gitops":               flux.Install,
	"git-operator":         gitoperator.Install,
	"harbor":               harbor.Deploy,
	"istio-operator":       istiooperator.Install,
	"journalbeat":          journalbeat.Deploy,
	"karina-operator":      karinaoperator.Install,
	"kpack":                kpack.Deploy,
	"kube-web-view":        kubewebview.Install,
	"kube-resource-report": kuberesourcereport.Install,
	"logs-exporter":        logsexporter.Install,
	"minio":                minio.Install,
	"monitoring":           monitoring.Install,
	"ingress":              ingress.Install,
	"opa":                  opa.Install,
	"nsx":                  nsx.Install,
	"packetbeat":           packetbeat.Deploy,
	"redis-operator":       redisoperator.Install,
	"rabbitmq-operator":    rabbitmqoperator.Install,
	"postgres-operator":    postgresoperator.Deploy,
	"registry-creds":       registrycreds.Install,
	"s3-upload-cleaner":    s3uploadcleaner.Deploy,
	"sealed-secrets":       sealedsecrets.Install,
	"stubs":                stubs.Install,
	"tekton":               tekton.Install,
	"template-operator":    templateoperator.Install,
	"vault":                vault.Deploy,
	"velero":               velero.Install,
}

var PhasesExtra = map[string]DeployFn{
	"cert-manager":      certmanager.Install,
	"platform-operator": platformoperator.Install,
	"vsphere":           vsphere.Install,
	"quack":             quack.Install,
}

var PhaseOrder = []string{"pre", "crds", "calico", "antrea", "nsx", "base", "stubs", "minio", "postgres-operator", "dex", "vault"}

func GetPhases() map[string]DeployFn {
	res := map[string]DeployFn{}
	for k, v := range Phases {
		res[k] = v
	}
	return res
}
