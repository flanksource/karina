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
	"github.com/flanksource/karina/pkg/phases/dashboard"
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
	"github.com/flanksource/karina/pkg/phases/keptn"
	"github.com/flanksource/karina/pkg/phases/kiosk"
	"github.com/flanksource/karina/pkg/phases/konfigmanager"
	"github.com/flanksource/karina/pkg/phases/kpack"
	"github.com/flanksource/karina/pkg/phases/kuberesourcereport"
	"github.com/flanksource/karina/pkg/phases/kubewebview"
	"github.com/flanksource/karina/pkg/phases/logsexporter"
	"github.com/flanksource/karina/pkg/phases/minio"
	"github.com/flanksource/karina/pkg/phases/mongodboperator"
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

type Cmd string

const (
	Apacheds           Cmd = "apacheds"
	Antrea             Cmd = "antrea"
	ArgoRollouts       Cmd = "argo-rollouts"
	ArgoOperator       Cmd = "argo-operator"
	Auditbeat          Cmd = "auditbeat"
	Base               Cmd = "base"
	BootstrapCmd       Cmd = "bootstrap"
	Calico             Cmd = "calico"
	Canary             Cmd = "canary"
	CertManager        Cmd = "cert-manager"
	Cni                Cmd = "cni"
	CloudController    Cmd = "cloud-controller"
	ConfigmapReloader  Cmd = "configmap-reloader"
	Crds               Cmd = "crds"
	Csi                Cmd = "csi"
	Dashboard          Cmd = "dashboard"
	Dex                Cmd = "dex"
	Eck                Cmd = "eck"
	Elasticsearch      Cmd = "elasticsearch"
	Eventrouter        Cmd = "eventrouter"
	Externaldns        Cmd = "externaldns"
	Filebeat           Cmd = "filebeat"
	Flux               Cmd = "flux"
	GitOperator        Cmd = "git-operator"
	Gitops             Cmd = "gitops"
	Harbor             Cmd = "harbor"
	Ingress            Cmd = "ingress"
	IstioOperator      Cmd = "istio-operator"
	Journalbeat        Cmd = "journalbeat"
	Keptn              Cmd = "keptn"
	Kiosk              Cmd = "kiosk"
	KarinaOperator     Cmd = "karina-operator"
	KonfigManager      Cmd = "konfig-manager"
	Kpack              Cmd = "kpack"
	PlatformCmd        Cmd = "platform"
	PlatformOperator   Cmd = "platform-operator"
	KubeResourceReport Cmd = "kube-resource-report"
	KubeWebView        Cmd = "kube-web-view"
	LogsExporter       Cmd = "logs-exporter"
	Minio              Cmd = "minio"
	MongodbOperator    Cmd = "mongodb-operator"
	Monitoring         Cmd = "monitoring"
	NodeLocalDNS       Cmd = "node-local-dns"
	Nsx                Cmd = "nsx"
	Opa                Cmd = "opa"
	Packetbeat         Cmd = "packetbeat"
	PostgresOperator   Cmd = "postgres-operator"
	Pre                Cmd = "pre"
	Quack              Cmd = "quack"
	RabbitmqOperator   Cmd = "rabbitmq-operator"
	RedisOperator      Cmd = "redis-operator"
	RegistryCreds      Cmd = "registry-creds"
	S3UploadCleaner    Cmd = "s3-upload-cleaner"
	SealedSecrets      Cmd = "sealed-secrets"
	StubsCmd           Cmd = "stubs"
	Tekton             Cmd = "tekton"
	TemplateOperator   Cmd = "template-operator"
	Vault              Cmd = "vault"
	Velero             Cmd = "velero"
	Vsphere            Cmd = "vsphere"
)

type Phase struct {
	Fn   DeployFn
	Name Cmd
}

func makePhase(df DeployFn, name Cmd) Phase {
	return Phase{Fn: df, Name: name}
}

var Phases = map[Cmd]Phase{
	ArgoRollouts:       makePhase(argorollouts.Deploy, ArgoRollouts),
	ArgoOperator:       makePhase(argocdoperator.Deploy, ArgoOperator),
	Auditbeat:          makePhase(auditbeat.Deploy, Auditbeat),
	Canary:             makePhase(canary.Deploy, Canary),
	Eck:                makePhase(eck.Deploy, Eck),
	Elasticsearch:      makePhase(elasticsearch.Deploy, Elasticsearch),
	Eventrouter:        makePhase(eventrouter.Deploy, Eventrouter),
	Externaldns:        makePhase(externaldns.Install, Externaldns),
	Dashboard:          makePhase(dashboard.Install, Dashboard),
	Filebeat:           makePhase(filebeat.Deploy, Filebeat),
	Flux:               makePhase(flux.InstallV2, Flux),
	GitOperator:        makePhase(gitoperator.Install, GitOperator),
	Gitops:             makePhase(flux.Install, Gitops),
	Harbor:             makePhase(harbor.Deploy, Harbor),
	IstioOperator:      makePhase(istiooperator.Install, IstioOperator),
	Journalbeat:        makePhase(journalbeat.Deploy, Journalbeat),
	Keptn:              makePhase(keptn.Deploy, Keptn),
	Kiosk:              makePhase(kiosk.Deploy, Kiosk),
	KarinaOperator:     makePhase(karinaoperator.Install, KarinaOperator),
	KonfigManager:      makePhase(konfigmanager.Deploy, KonfigManager),
	Kpack:              makePhase(kpack.Deploy, Kpack),
	PlatformCmd:        makePhase(Platform, PlatformCmd),
	KubeResourceReport: makePhase(kuberesourcereport.Install, KubeResourceReport),
	KubeWebView:        makePhase(kubewebview.Install, KubeWebView),
	LogsExporter:       makePhase(logsexporter.Install, LogsExporter),
	MongodbOperator:    makePhase(mongodboperator.Deploy, MongodbOperator),
	Monitoring:         makePhase(monitoring.Install, Monitoring),
	Opa:                makePhase(opa.Install, Opa),
	Packetbeat:         makePhase(packetbeat.Deploy, Packetbeat),
	RabbitmqOperator:   makePhase(rabbitmqoperator.Install, RabbitmqOperator),
	RedisOperator:      makePhase(redisoperator.Install, RedisOperator),
	RegistryCreds:      makePhase(registrycreds.Install, RegistryCreds),
	S3UploadCleaner:    makePhase(s3uploadcleaner.Deploy, S3UploadCleaner),
	SealedSecrets:      makePhase(sealedsecrets.Install, SealedSecrets),
	Tekton:             makePhase(tekton.Install, Tekton),
	Vault:              makePhase(vault.Deploy, Vault),
	Velero:             makePhase(velero.Install, Velero),
}

var PhaseOrder = []Cmd{BootstrapCmd, Crds, Cni, Csi, CloudController, PlatformCmd}
var Bootstrap = compose(pre.Install, crds.Install, CNI, CSI, base.Install, Cloud, certmanager.Install, ingress.Install, quack.Install, minio.Install, templateoperator.Install, postgresoperator.Deploy, dex.Install)
var BootstrapPhases = []Cmd{Pre, Crds, Cni, Csi, Base, CloudController, CertManager, Ingress, Quack, Minio, TemplateOperator, PostgresOperator, Dex}
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

var PhasesExtra = map[Cmd]Phase{
	Apacheds:          makePhase(apacheds.Install, Apacheds),
	Antrea:            makePhase(antrea.Install, Antrea),
	Base:              makePhase(base.Install, Base),
	BootstrapCmd:      makePhase(Bootstrap, BootstrapCmd),
	Calico:            makePhase(calico.Install, Calico),
	CertManager:       makePhase(certmanager.Install, CertManager),
	Cni:               makePhase(CNI, Cni),
	ConfigmapReloader: makePhase(configmapreloader.Deploy, ConfigmapReloader),
	Crds:              makePhase(crds.Install, Crds),
	Csi:               makePhase(CSI, Csi),
	Dex:               makePhase(dex.Install, Dex),
	Ingress:           makePhase(ingress.Install, Ingress),
	Kiosk:             makePhase(kiosk.Deploy, Kiosk),
	Minio:             makePhase(minio.Install, Minio),
	NodeLocalDNS:      makePhase(nodelocaldns.Install, NodeLocalDNS),
	Nsx:               makePhase(nsx.Install, Nsx),
	PostgresOperator:  makePhase(postgresoperator.Deploy, PostgresOperator),
	PlatformOperator:  makePhase(platformoperator.Install, PlatformOperator),
	Pre:               makePhase(pre.Install, Pre),
	Quack:             makePhase(quack.Install, Quack),
	TemplateOperator:  makePhase(templateoperator.Install, TemplateOperator),
	Vsphere:           makePhase(vsphere.Install, Vsphere),
	CloudController:   makePhase(Cloud, CloudController),
	StubsCmd:          makePhase(Stubs, StubsCmd),
}

func mergePhases(phaseMaps ...map[Cmd]Phase) map[Cmd]Phase {
	res := map[Cmd]Phase{}
	for _, phaseMap := range phaseMaps {
		for k, v := range phaseMap {
			res[k] = v
		}
	}
	return res
}

func GetAllPhases() map[Cmd]Phase {
	return mergePhases(Phases, PhasesExtra)
}

func GetPhases() map[Cmd]Phase {
	return mergePhases(Phases)
}
