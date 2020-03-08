module github.com/moshloop/platform-cli

go 1.12

require (
<<<<<<< HEAD
	cloud.google.com/go v0.47.0 // indirect
	cloud.google.com/go/bigquery v1.2.0 // indirect
	cloud.google.com/go/storage v1.2.1 // indirect
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/aws/aws-sdk-go v1.15.78
	github.com/crunchydata/postgres-operator v4.2.1+incompatible
=======
	github.com/aws/aws-sdk-go v1.29.5
	github.com/davecgh/go-spew v1.1.1
>>>>>>> upstream/master
	github.com/dghubble/sling v1.3.0
	github.com/fatih/structs v1.1.0
<<<<<<< HEAD
	github.com/flanksource/commons v1.0.2
	github.com/gobuffalo/packr/v2 v2.7.1 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
=======
	github.com/flanksource/commons v1.1.1
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gogo/protobuf v1.3.1 // indirect
>>>>>>> upstream/master
	github.com/google/uuid v1.1.1
	github.com/grafana-tools/sdk v0.0.0-20191214173017-690a0c6bec7b
	github.com/imdario/mergo v0.3.6
	github.com/miekg/dns v1.1.22
	github.com/minio/minio-go/v6 v6.0.44
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mjibson/esc v0.2.0 // indirect
	github.com/moshloop/konfigadm v0.4.6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.1 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.5
<<<<<<< HEAD
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.20.2
	github.com/vmware/kube-fluentd-operator v1.11.0
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6
	golang.org/x/net v0.0.0-20191108063844-7e6e90b9ea88 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191107235519-f7ea15e60b12 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5

	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/cluster-bootstrap v0.0.0-20190831080953-99cb41cb5d35
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
=======
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.21.0
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/cluster-bootstrap v0.17.3
	sigs.k8s.io/kind v0.7.0
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
>>>>>>> upstream/master
)

replace (
	github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
	golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115
	gopkg.in/hairyhenderson/yaml.v2 => github.com/maxaudron/yaml v0.0.0-20190411130442-27c13492fe3c
	k8s.io/client-go => k8s.io/client-go v0.17.3
)
