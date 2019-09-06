module github.com/moshloop/platform-cli

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/cpuguy83/go-md2man v1.0.9-0.20180619205630-691ee98543af // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-getter v1.3.0
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.6
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moshloop/commons v0.0.0-20190826175941-22006c29e3f5
	github.com/moshloop/konfigadm v0.3.6
	github.com/pkg/errors v0.8.1
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/vmware/govmomi v0.20.2
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190830142957-1e83adbbebd0 // indirect
	google.golang.org/api v0.9.0 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	google.golang.org/grpc v1.23.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v0.0.0-20190828235140-8248d0a0e61a
	k8s.io/cluster-bootstrap v0.0.0-20190831080953-99cb41cb5d35
)

replace github.com/moshloop/commons => ../commons
