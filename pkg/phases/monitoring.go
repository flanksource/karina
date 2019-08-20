package phases

import (
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

func Monitoring(cfg types.PlatformConfig) error {
	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		if err := utils.Exec(".bin/gomplate --input-dir \"%s\" --output-dir build -c \".=%s\"", spec, cfg.Source); err != nil {
			return err
		}
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir("build"); err != nil {
		return err
	}

	log.Infoln("Building monitoring stack")
	if !utils.FileExists("jsonnetfile.json") {
		log.Infof("Initializing jsonnet bundler")
		if err := utils.Exec("../.bin/jb init"); err != nil {
			return err
		}
	}

	if !utils.FileExists("vendor/kube-prometheus") {
		log.Infof("Installing kube-prometheus")
		if err := utils.Exec("../.bin/jb install github.com/coreos/kube-prometheus/jsonnet/kube-prometheus"); err != nil {
			return err
		}
	}

	if err := utils.Exec("../.bin/jsonnet -J vendor -m . %s | xargs -I{} sh -c 'cat {} | ../.bin/gojsontoyaml > {}.yaml; rm -f {}' -- {}", "monitoring.jsonnet"); err != nil {
		return err
	}

	return nil
}
