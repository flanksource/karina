package phases

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
)

var manifests = []string{
	"https://docs.projectcalico.org/v3.8/manifests/calico.yaml",
	"https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.9/deploy/local-path-storage.yaml",
	"https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml",
	"https://raw.githubusercontent.com/heptiolabs/eventrouter/master/yaml/eventrouter.yaml",
	"https://download.elastic.co/downloads/eck/0.9.0/all-in-one.yaml",
}

func Build(cfg types.PlatformConfig) error {
	tmp, _ := ioutil.TempFile("", "config*.yml")
	data, _ := yaml.Marshal(cfg)
	tmp.WriteString(string(data))
	os.Mkdir("build", 0750)
	for _, manifest := range manifests {
		name := path.Base(manifest)
		dest := "build/" + name
		if utils.FileExists(dest) {
			continue
		}
		log.Infof("Downloading %s\n", manifest)
		body, err := utils.GET(manifest)
		if err != nil {
			return err
		}
		ioutil.WriteFile(dest, body, 0644)
	}

	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		if err := utils.Exec(".bin/gomplate --input-dir \"%s\" --output-dir build -c \".=%s\"", spec, tmp.Name()); err != nil {
			return err
		}
	}

	if utils.FileExists("kustomization.yaml") {
		log.Infoln("Building with kustomize")
		os.Remove("build/kustomize.yml")
		if err := utils.Exec(".bin/kustomize build > build/kustomize.yml"); err != nil {
			return err
		}
	}

	return nil
}
