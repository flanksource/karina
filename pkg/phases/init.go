package phases

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
)

var manifests = []string{
	"https://docs.projectcalico.org/v3.8/manifests/calico.yaml",
	"https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/mandatory.yaml",
	"https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/baremetal/service-nodeport.yaml",
	"https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml",
	"https://raw.githubusercontent.com/heptiolabs/eventrouter/master/yaml/eventrouter.yaml",
}

func Init(cfg types.PlatformConfig) error {
	os.Mkdir("build", 0750)
	for _, manifest := range manifests {
		name := path.Base(manifest)
		body, err := utils.GET(manifest)
		if err != nil {
			return err
		}
		ioutil.WriteFile("build/"+name, body, 0644)
	}
	return nil
}
