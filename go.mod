module github.com/flanksource/karina

go 1.16

require (
	cloud.google.com/go/storage v1.8.0 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/aws/aws-sdk-go v1.34.30
	github.com/blang/semver/v4 v4.0.0
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/coreos/prometheus-operator v0.37.0
	github.com/dghubble/sling v1.3.0
	github.com/fatih/structs v1.1.0
	github.com/flanksource/commons v1.5.2
	github.com/flanksource/kommons v0.10.0
	github.com/flanksource/konfigadm v0.6.0-2-g9751ff1
	github.com/flanksource/template-operator-library v0.1.0
	github.com/frankban/quicktest v1.8.1 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0 // indirect
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/go-pg/pg/v9 v9.1.6
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.4.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11
	github.com/jetstack/cert-manager v1.2.0
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/kr/pretty v0.2.1
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/miekg/dns v1.1.31
	github.com/minio/minio-go/v6 v6.0.44
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olivere/elastic/v7 v7.0.13
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.10.0
	github.com/sirupsen/logrus v1.7.1
	github.com/spf13/cobra v1.1.1
	github.com/thoas/go-funk v0.7.0
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/vbauerster/mpb/v5 v5.0.3
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.21.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.mozilla.org/sops/v3 v3.6.1
	go.opencensus.io v0.22.4 // indirect
	go.uber.org/zap v1.15.0
	gocloud.dev v0.19.0 // indirect
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/api v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/flanksource/yaml.v3 v3.1.1
	gopkg.in/ini.v1 v1.56.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.17.2
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/kind v0.10.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/flanksource/kommons => github.com/tobernguyen/kommons v0.3.5-0.20210409051250-d6121dbc87ce
	github.com/flanksource/template-operator-library => github.com/tobernguyen/template-operator-library v0.1.1-0.20210408061816-57964c040f68
	github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
	golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115
	gopkg.in/hairyhenderson/yaml.v2 => github.com/maxaudron/yaml v0.0.0-20190411130442-27c13492fe3c
	k8s.io/client-go => k8s.io/client-go v0.20.4
)
