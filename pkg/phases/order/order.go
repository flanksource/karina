package order

import (
	"github.com/flanksource/karina/pkg/phases/antrea"
	"github.com/flanksource/karina/pkg/phases/apacheds"
	"github.com/flanksource/karina/pkg/phases/argocdoperator"
	"github.com/flanksource/karina/pkg/phases/argorollouts"
	"github.com/flanksource/karina/pkg/phases/auditbeat"
	"github.com/flanksource/karina/pkg/phases/base"
	"github.com/flanksource/karina/pkg/phases/calico"
	"github.com/flanksource/karina/pkg/phases/canary"
	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/phases/configmapreloader"
	"github.com/flanksource/karina/pkg/phases/crds"
	"github.com/flanksource/karina/pkg/phases/csi/localpath"
	"github.com/flanksource/karina/pkg/phases/csi/nfs"
	"github.com/flanksource/karina/pkg/phases/csi/s3"
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
	"github.com/flanksource/karina/pkg/phases/nodelocaldns"
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
	"github.com/flanksource/karina/pkg/phases/tekton"
	"github.com/flanksource/karina/pkg/phases/templateoperator"
	"github.com/flanksource/karina/pkg/phases/vault"
	"github.com/flanksource/karina/pkg/phases/velero"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"
)

type DeployFn func(p *platform.Platform) error

var Phases = map[string]DeployFn{
	"argo-rollouts":        argorollouts.Deploy,
	"argocd-operator":      argocdoperator.Deploy,
	"auditbeat":            auditbeat.Deploy,
	"canary":               canary.Deploy,
	"eck":                  eck.Deploy,
	"elasticsearch":        elasticsearch.Deploy,
	"eventrouter":          eventrouter.Deploy,
	"externaldns":          externaldns.Install,
	"filebeat":             filebeat.Deploy,
	"git-operator":         gitoperator.Install,
	"gitops":               flux.Install,
	"harbor":               harbor.Deploy,
	"istio-operator":       istiooperator.Install,
	"journalbeat":          journalbeat.Deploy,
	"karina-operator":      karinaoperator.Install,
	"kpack":                kpack.Deploy,
	"platform":             Platform,
	"kube-resource-report": kuberesourcereport.Install,
	"kube-web-view":        kubewebview.Install,
	"logs-exporter":        logsexporter.Install,
	"monitoring":           monitoring.Install,
	"opa":                  opa.Install,
	"packetbeat":           packetbeat.Deploy,
	"rabbitmq-operator":    rabbitmqoperator.Install,
	"redis-operator":       redisoperator.Install,
	"registry-creds":       registrycreds.Install,
	"s3-upload-cleaner":    s3uploadcleaner.Deploy,
	"sealed-secrets":       sealedsecrets.Install,
	"tekton":               tekton.Install,
	"velero":               velero.Install,
	"vault":                vault.Deploy,
}

var PhaseOrder = []string{"bootstrap", "crds", "cni", "csi", "cloud", "platform"}
var Bootstrap = compose(pre.Install, crds.Install, CNI, CSI, base.Install, Cloud, certmanager.Install, ingress.Install, quack.Install, minio.Install, templateoperator.Install, postgresoperator.Deploy, dex.Install)
var CSI = compose(localpath.Install, s3.Install, nfs.Install)
var CNI = compose(calico.Install, antrea.Install, nsx.Install, nodelocaldns.Install)
var Cloud = compose(vsphere.Install)
var Platform = compose(platformoperator.Install, kiosk.Deploy, configmapreloader.Deploy)
var Stubs = compose(minio.Install, apacheds.Install)

func compose(fns ...DeployFn) DeployFn {
	return func(p *platform.Platform) error {
		for _, DeployFn := range fns {
			if err := DeployFn(p); err != nil {
				return err
			}
		}
		return nil
	}
}

var PhasesExtra = map[string]DeployFn{
	"apacheds":           apacheds.Install,
	"antrea":             antrea.Install,
	"base":               base.Install,
	"bootstrap":          Bootstrap,
	"calico":             calico.Install,
	"cert-manager":       certmanager.Install,
	"cni":                CNI,
	"configmap-reloader": configmapreloader.Deploy,
	"crds":               crds.Install,
	"csi":                CSI,
	"dex":                dex.Install,
	"ingress":            ingress.Install,
	"kiosk":              kiosk.Deploy,
	"minio":              minio.Install,
	"node-local-dns":     nodelocaldns.Install,
	"nsx":                nsx.Install,
	"postgres-operator":  postgresoperator.Deploy,
	"platform-operator":  platformoperator.Install,
	"pre":                pre.Install,
	"quack":              quack.Install,
	"template-operator":  templateoperator.Install,
	"vsphere":            vsphere.Install,
	"cloud-controller":   Cloud,
	"stubs":              Stubs,
}

func GetAllPhases() map[string]DeployFn {
	res := map[string]DeployFn{}
	for k, v := range Phases {
		res[k] = v
	}
	for k, v := range PhasesExtra {
		res[k] = v
	}
	return res
}

func GetPhases() map[string]DeployFn {
	res := map[string]DeployFn{}
	for k, v := range Phases {
		res[k] = v
	}
	return res
}
