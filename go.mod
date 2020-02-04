module github.com/moshloop/platform-cli

go 1.12

require (
	cloud.google.com/go v0.47.0 // indirect
	cloud.google.com/go/bigquery v1.2.0 // indirect
	cloud.google.com/go/storage v1.2.1 // indirect
	github.com/aws/aws-sdk-go v1.16.26
	github.com/dghubble/sling v1.3.0
	github.com/fatih/structs v1.1.0
	github.com/flanksource/commons v1.0.2
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/grafana-tools/sdk v0.0.0-20191214173017-690a0c6bec7b
	github.com/imdario/mergo v0.3.6
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/miekg/dns v1.1.22
	github.com/minio/minio-go/v6 v6.0.44
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moshloop/konfigadm v0.4.6
	github.com/pkg/errors v0.9.0
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/vmware/go-vmware-nsxt v0.0.0-20190201205556-16aa0443042d
	github.com/vmware/govmomi v0.20.3
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6 // indirect
	golang.org/x/net v0.0.0-20191108063844-7e6e90b9ea88 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.0 // indirect

	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/cluster-bootstrap v0.17.0
	k8s.io/kubernetes v1.17.0
	sigs.k8s.io/kind v0.7.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
	golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115

	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	k8s.io/apiserver => k8s.io/apiserver v0.17.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.0
	k8s.io/code-generator => k8s.io/code-generator v0.17.0
	k8s.io/component-base => k8s.io/component-base v0.17.0
	k8s.io/cri-api => k8s.io/cri-api v0.17.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.0
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.0
	k8s.io/kubectl => k8s.io/kubectl v0.17.0
	k8s.io/kubelet => k8s.io/kubelet v0.17.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.0
	k8s.io/metrics => k8s.io/metrics v0.17.0
	k8s.io/node-api => k8s.io/node-api v0.17.0
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.0
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.0
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.0
)
