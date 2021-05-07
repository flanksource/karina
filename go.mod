module github.com/flanksource/karina

go 1.16

require (
	github.com/aws/aws-sdk-go v1.34.30
	github.com/blang/semver/v4 v4.0.0
	github.com/coreos/prometheus-operator v0.37.0
	github.com/dghubble/sling v1.3.0
	github.com/fatih/structs v1.1.0
	github.com/flanksource/commons v1.5.6
	github.com/flanksource/kommons v0.19.0
	github.com/flanksource/kommons/testenv v0.0.0-20210628131214-7dac6b472d63
	github.com/flanksource/konfigadm v0.11.0
	github.com/flanksource/template-operator-library v0.1.6
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11
	github.com/jackc/pgx/v4 v4.11.0
	github.com/jetstack/cert-manager v1.2.0
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/kr/pretty v0.2.1
	github.com/miekg/dns v1.1.31
	github.com/minio/minio-go/v6 v6.0.44
	github.com/olivere/elastic/v7 v7.0.13
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.10.0
	github.com/sirupsen/logrus v1.7.1
	github.com/spf13/cobra v1.1.1
	github.com/thoas/go-funk v0.7.0
	github.com/vbauerster/mpb/v5 v5.0.3
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.21.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.mozilla.org/sops/v3 v3.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/flanksource/yaml.v3 v3.1.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.17.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/kind v0.10.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
	golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115
	gopkg.in/hairyhenderson/yaml.v2 => github.com/maxaudron/yaml v0.0.0-20190411130442-27c13492fe3c
	k8s.io/client-go => k8s.io/client-go v0.20.4
)
