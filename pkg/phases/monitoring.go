package phases

import (
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

func Monitoring(cfg types.PlatformConfig) error {
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir("build"); err != nil {
		return err
	}

	if cfg.BuildOptions.Monitoring {
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
	}

	return nil
}
