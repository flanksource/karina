package monitoring

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/commons/exec"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/utils"
)

func Install(p *platform.Platform) error {

	if p.Monitoring == nil || p.Monitoring.Disabled {
		return nil
	}

	jb := p.GetBinary("jb")
	jsonnet := p.GetBinary("jsonnet")
	kubectl := p.GetKubectl()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Mkdir("build", 0644)
	if err := os.Chdir("build"); err != nil {
		return err
	}

	spec, err := p.Template("monitoring.jsonnet")
	if err != nil {
		return err
	}
	alerts, err := p.Template("alertmanager-config.yaml.tpl")
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
		if err := jb("install github.com/coreos/kube-prometheus/jsonnet/kube-prometheus@%s", p.Monitoring.Version); err != nil {
			return err
		}
	}

	os.Mkdir("monitoring", 0755)
	if err := jsonnet("-J vendor -m monitoring monitoring.jsonnet"); err != nil {
		return err
	}

	//jsonnet outputs files without an extension, so we add a json extension
	exec.SafeExec("bash -c 'for f in $(ls monitoring); do mv monitoring/$f monitoring/$f.json; done'")

	// run kubectl once to ensure CRD's are created
	kubectl("apply -f monitoring/")
	return kubectl("apply -f monitoring/")
}
