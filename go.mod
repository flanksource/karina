module github.com/moshloop/platform-cli

go 1.12

require (
	cloud.google.com/go v0.47.0 // indirect
	cloud.google.com/go/bigquery v1.2.0 // indirect
	cloud.google.com/go/storage v1.2.1 // indirect
	github.com/apex/log v1.1.0
	github.com/dghubble/sling v1.3.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/grafana-tools/sdk v0.0.0-20191214173017-690a0c6bec7b
	github.com/imdario/mergo v0.3.6
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/miekg/dns v1.1.22
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/moshloop/commons v0.0.3-0.20191104124838-faba5f315c26
	github.com/moshloop/konfigadm v0.4.6
	github.com/pkg/errors v0.8.1
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vmware/govmomi v0.20.2
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6 // indirect
	golang.org/x/net v0.0.0-20191108063844-7e6e90b9ea88 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191107235519-f7ea15e60b12 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/cluster-bootstrap v0.0.0-20190831080953-99cb41cb5d35
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)

// replace github.com/moshloop/commons => ../commons
// replace github.com/moshloop/konfigadm => ../konfigadm
replace github.com/go-check/check v1.0.0-20180628173108-788fd7840127 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15

// latest golang.org/x/exp has a dependency on dmitri.shuralyov.com which is down
replace golang.org/x/exp => golang.org/x/exp v0.0.0-20190829150108-63fe5bdad115

replace github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
