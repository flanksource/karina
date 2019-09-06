package phases

import (
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/files"
	"github.com/moshloop/commons/is"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func Build(cfg types.PlatformConfig) error {
	tmp, _ := ioutil.TempFile("", "config*.yml")
	data, _ := yaml.Marshal(cfg)
	tmp.WriteString(string(data))
	os.Mkdir("build", 0755)
	gomplate := deps.Binary("gomplate", cfg.Versions["gomplate"], ".bin")
	kustomize := deps.Binary("kustomize", cfg.Versions["kustomize"], ".bin")
	// for _, manifest := range manifests {
	// 	name := path.Base(manifest)
	// 	dest := "build/" + name
	// 	if utils.FileExists(dest) {
	// 		continue
	// 	}
	// 	log.Infof("Downloading %s\n", manifest)
	// 	body, err := utils.GET(manifest)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	ioutil.WriteFile(dest, body, 0644)
	// }

	for k, url := range cfg.Resources {
		if !is.File(k) {
			if err := files.Getter(url, k); err != nil {
				return err
			}
		}
	}

	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		absPath, _ := os.Readlink(spec)
		if err := gomplate("--input-dir \"%s\" --output-dir build -c \".=%s\"", absPath, tmp.Name()); err != nil {
			return err
		}
	}

	if utils.FileExists("kustomization.yaml") {
		log.Infoln("Building with kustomize")
		os.Remove("build/kustomize.yml")
		if err := kustomize("build > build/kustomize.yml"); err != nil {
			return err
		}
	}

	return nil
}
