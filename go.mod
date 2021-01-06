module github.com/flanksource/karina

go 1.12

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/aws/aws-sdk-go v1.29.25
	github.com/blang/semver/v4 v4.0.0
	github.com/coreos/prometheus-operator v0.37.0
	github.com/dghubble/sling v1.3.0
	github.com/fatih/structs v1.1.0
	github.com/flanksource/commons v1.4.3
	github.com/flanksource/kommons v0.1.7
	github.com/flanksource/konfigadm v0.6.0-2-g9751ff1
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0
	github.com/go-pg/pg/v9 v9.1.6
	github.com/go-test/deep v1.0.7
	github.com/google/uuid v1.1.1
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.9
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/miekg/dns v1.1.22
	github.com/minio/minio-go/v6 v6.0.44
	github.com/mitchellh/mapstructure v1.3.3
	github.com/olivere/elastic/v7 v7.0.13
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/common v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/vbauerster/mpb/v5 v5.0.3
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.21.0
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.27.0
	gopkg.in/flanksource/yaml.v3 v3.1.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.17.2
	k8s.io/klog/v2 v2.4.0
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/kind v0.7.1-0.20200303021537-981bd80d3802
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
	golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191230161307-f3c370f40bfb
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	gopkg.in/hairyhenderson/yaml.v2 => github.com/maxaudron/yaml v0.0.0-20190411130442-27c13492fe3c
	k8s.io/client-go => k8s.io/client-go v0.19.3
)
