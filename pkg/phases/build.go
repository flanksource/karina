package phases

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flanksource/yaml"

	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/is"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
)

func Build(cfg types.PlatformConfig) error {
	tmp, _ := ioutil.TempFile("", "config*.yaml")
	data, _ := yaml.Marshal(cfg)
	tmp.WriteString(string(data)) // nolint: errcheck
	os.Mkdir("build", 0755)       // nolint: errcheck
	gomplate := deps.Binary("gomplate", cfg.Versions["gomplate"], ".bin")
	kustomize := deps.Binary("kustomize", cfg.Versions["kustomize"], ".bin")

	for k, url := range cfg.Resources {
		if !is.File(k) {
			if err := files.Getter(url, k); err != nil {
				return fmt.Errorf("build: failed get url: %v", err)
			}
		}
	}

	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		absPath, _ := os.Readlink(spec)
		if err := gomplate("--input-dir \"%s\" --output-dir build -c \".=%s\"", absPath, tmp.Name()); err != nil {
			return fmt.Errorf("build: failed to template: %v", err)
		}
	}

	if files.Exists("kustomization.yaml") {
		log.Infoln("Building with kustomize")
		os.Remove("build/kustomize.yaml")
		if err := kustomize("build > build/kustomize.yaml"); err != nil {
			return fmt.Errorf("build: failed to apply kustomize: %v", err)
		}
	}

	return nil
}
