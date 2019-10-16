package monitoring

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/commons/deps"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/utils"
)

func Install(cfg *platform.Platform) error {

	jb := deps.Binary("jb", cfg.Versions["jb"], ".bin")
	jsonnet := deps.Binary("jsonnet", cfg.Versions["jsonnet"], ".bin")

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir("build"); err != nil {
		return err
	}

	spec, err := cfg.Template("monitoring.jsonnet")
	alerts, err := cfg.Template("alertmanager-config.yaml.tpl")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile("monitoring.jsonnet", []byte(spec), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile("alertmanager-config.yaml.tpl", []byte(alerts), 0644); err != nil {
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
		if err := jb("install github.com/coreos/kube-prometheus/jsonnet/kube-prometheus@%s", cfg.Monitoring.Version); err != nil {
			return err
		}
	}

	return jsonnet("-J vendor -m . %s", "monitoring.jsonnet")
}
