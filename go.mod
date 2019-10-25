module github.com/moshloop/platform-cli

go 1.12

require (
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/dghubble/sling v1.3.0
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobuffalo/logger v1.0.1 // indirect
	github.com/gobuffalo/packd v0.3.0
	github.com/gobuffalo/packr/v2 v2.6.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.6
	github.com/miekg/dns v1.1.22
	github.com/moshloop/commons v0.0.3-0.20191025113427-254e5e699d44
	github.com/moshloop/konfigadm v0.3.6
	github.com/pkg/errors v0.8.1
	github.com/rogpeppe/go-internal v1.4.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vmware/govmomi v0.20.2
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/exp v0.0.0-20190919035709-81c71964d733 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	golang.org/x/tools v0.0.0-20190925020647-22afafe3322a // indirect
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/appengine v1.6.3 // indirect
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731 // indirect
	google.golang.org/grpc v1.23.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/cluster-bootstrap v0.0.0-20190831080953-99cb41cb5d35
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)

// replace github.com/moshloop/commons => ../commons
// replace github.com/moshloop/konfigadm => ../konfigadm

replace github.com/russross/blackfriday v2.0.0+incompatible => github.com/russross/blackfriday v1.5.2
