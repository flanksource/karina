package monitoring

import (
	"os"

	"github.com/moshloop/commons/deps"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func Install(cfg *platform.Platform) error {

	jb := deps.Binary("jb", cfg.Versions["jb"], ".bin")
	jsonnet := deps.Binary("jsonnet", cfg.Versions["jsonnet"], ".bin")
	gomplate := deps.Binary("gomplate", cfg.Versions["gomplate"], ".bin")
	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		if err := gomplate("--input-dir \"%s\" --output-dir build -c \".=%s\"", spec, cfg.Source); err != nil {
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
		if err := jb("init"); err != nil {
			return err
		}
	}

	if !utils.FileExists("vendor/kube-prometheus") {
		log.Infof("Installing kube-prometheus")
		if err := jb("install github.com/coreos/kube-prometheus/jsonnet/kube-prometheus"); err != nil {
			return err
		}
	}

	return jsonnet("-J vendor -m . %s | xargs -I{} sh -c 'cat {} | ../.bin/gojsontoyaml > {}.yaml; rm -f {}' -- {}", "monitoring.jsonnet")
}
