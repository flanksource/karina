package pgo

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/exec"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	PGO             = "pgo"
	PGOUSER         = "PGOUSER"
	PGOPASS         = "PGOPASS"
	PGO_CLIENT_CERT = "PGO_CLIENT_CERT"
	PGO_CLIENT_KEY  = "PGO_CLIENT_KEY"
	CO_CLIENT_CERT  = "CO_CLIENT_CERT"
	CO_CLIENT_KEY   = "CO_CLIENT_KEY"
	CO_CA_CERT      = "CO_CA_CERT"
	PGO_CA_CERT     = "PGO_CA_CERT"
)

func getEnvMap(p *platform.Platform) map[string]string {
	kubeconfig, _ := p.GetKubeConfig()
	home := os.Getenv("HOME") + "/.pgouser-" + p.Name
	crt := home + ".crt"
	key := home + ".key"

	return map[string]string{
		"PATH":       ".bin:" + os.Getenv("PATH"),
		"KUBECONFIG": kubeconfig,

		// 4.0.0 vars
		"PGO_OPERATOR_NAMESPACE": PGO,
		"NAMESPACE":              "pgo",
		"PGO_APISERVER_URL":      fmt.Sprintf("https://postgres-operator.%s", p.Domain),
		"PGO_CMD":                ".bin/kubectl",
		"PGOROOT":                "build/pgo",
		"PGO_IMAGE_PREFIX":       "crunchydata",
		"PGO_BASEOS":             "centos7",
		"PGO_INSTALLATION_NAME":  "pgo",
		"PGO_NAMESPACE":          "pgo",
		"PGO_VERSION":            strings.ReplaceAll(p.PGO.Version, "v", ""),
		"PGO_IMAGE_TAG":          "centos7-" + strings.ReplaceAll(p.PGO.Version, "v", ""),
		"PGO_CA_CERT":            crt,
		"PGO_CLIENT_CERT":        crt,
		"PGO_CLIENT_KEY":         key,

		// 3.5.4 vars
		"CO_IMAGE_PREFIX":  "crunchydata",
		"CO_CMD":           "kubectl",
		"CO_UI":            "false",
		"CO_NAMESPACE":     PGO,
		"COROOT":           "build/pgo",
		"CO_IMAGE_TAG":     "centos7-" + strings.ReplaceAll(p.PGO.Version, "v", ""),
		"CO_APISERVER_URL": fmt.Sprintf("https://postgres-operator.%s", p.Domain),
		"CO_CA_CERT":       crt,
		"CO_CLIENT_CERT":   crt,
		"CO_CLIENT_KEY":    key,
		"PGOUSER":          home,
	}
}

func getPgoAuth(p *platform.Platform) (user, pass string) {
	if secret := p.GetSecret("pgo", "pgouser-pgoadmin"); secret != nil {
		user = string((*secret)["username"])
		pass = string((*secret)["password"])
		return
	}
	if secret := p.GetSecret("pgo", "pgo-auth-secret"); secret != nil {
		pgouser := string((*secret)["pgouser"])
		user = strings.Split(pgouser, ":")[0]
		pass = strings.Split(pgouser, ":")[1]
	}
	return
}

func getEnv(p *platform.Platform) (*map[string]string, error) {

	ENV := getEnvMap(p)
	home := ENV["PGOUSER"]
	user, pass := getPgoAuth(p)
	passwd := fmt.Sprintf("%s:%s", user, pass)
	log.Debugf("Writing %s", home)
	if err := ioutil.WriteFile(home, []byte(passwd), 0644); err != nil {
		log.Tracef("getEnv: Failed to write file: %s", err)
		return nil, err
	}

	secrets := *p.GetSecret("pgo", "pgo.tls")

	log.Debugf("Writing %s", ENV["PGO_CLIENT_CERT"])
	if err := ioutil.WriteFile(ENV["PGO_CLIENT_CERT"], secrets["tls.crt"], 0644); err != nil {
		log.Tracef("getEnv: Failed to write file: %s", err)
		return nil, err
	}

	log.Debugf("Writing %s", ENV["PGO_CLIENT_KEY"])
	if err := ioutil.WriteFile(ENV["PGO_CLIENT_KEY"], secrets["tls.key"], 0644); err != nil {
		log.Tracef("getEnv: Failed to write file: %s", err)
		return nil, err
	}

	return &ENV, nil
}

func ClientSetup(p *platform.Platform) error {
	if p.PGO == nil || p.PGO.Disabled {
		return nil
	}
	ENV, err := getEnv(p)
	if err != nil {
		log.Tracef("ClientSetup: Failed to get env: %s", err)
		return err
	}

	if p.DryRun {
		return nil
	}

	for k, v := range *ENV {
		fmt.Printf("export %s=%s\n", k, v)
	}
	deps.InstallDependency("pgo", getPGOTag(p.PGO.Version), ".bin")
	return nil
}

func Install(p *platform.Platform) error {
	if p.PGO == nil || p.PGO.Disabled {
		return nil
	}
	ENV := getEnvMap(p)

	for k, v := range ENV {
		log.Tracef("export %s=%s\n", k, v)
	}

	if files.Exists("build/pgo") {
		exec.Exec("cd build/pgo; git clean -fdx && git reset . ")
	}
	if err := files.Getter("git::https://github.com/CrunchyData/postgres-operator.git?ref="+getPGOTag(p.PGO.Version), "build/pgo"); err != nil {
		log.Tracef("Install: Failed to download pgo: %s", err)
		return err
	}

	if runtime.GOOS == "darwin" {
		// cp -R behavior seems to handle directories differently on macosx and linux?
		exec.ExecfWithEnv("cp -Rv overlays/pgo/ build/pgo", ENV)
	} else {
		exec.ExecfWithEnv("cp -Rv overlays/pgo/ build/", ENV)
	}
	template, err := p.Template("pgo.yaml", "templates")
	if err != nil {
		log.Warn(err)
	}
	templateFile := text.ToFile(template, ".yaml")
	exec.ExecfWithEnv(fmt.Sprintf("cp -v %s build/pgo/conf/postgres-operator/pgo.yaml", templateFile), ENV)
	if err := p.CreateOrUpdateNamespace(PGO, nil, nil); err != nil {
		log.Tracef("Install: Failed to create/update namespace: %s", err)
		return err
	}

	if err := p.ExposeIngressTLS("pgo", "postgres-operator", 8443); err != nil {
		log.Tracef("Install: Failed to expose ingress: %s", err)
		return err
	}

	if p.DryRun {
		return nil
	}
	if err := exec.ExecfWithEnv("/bin/bash  build/pgo/deploy/install-rbac.sh", ENV); err != nil {
		log.Tracef("Install: Failed to install rbac: %s", err)
		return err
	}
	return exec.ExecfWithEnv("/bin/bash build/pgo/deploy/deploy.sh", ENV)
}

func GetPGO(p *platform.Platform) (deps.BinaryFunc, error) {
	env, err := getEnv(p)
	if err != nil {
		log.Tracef("GetPGO: Failed to get env: %s", err)
		return nil, err
	}
	return deps.BinaryWithEnv(PGO, getPGOTag(p.PGO.Version), ".bin", *env), nil
}

// Takes version coming from config and returns correct git tag
// Strips/Adds "v" if required. Supports only 3.5.4, 4.0.0 and 4.2.0 versions
func getPGOTag(version string) string {
	var gitTag string
	if version == "4.0.0" || version == "3.5.4" {
		gitTag = strings.ReplaceAll(version, "v", "")
	} else if !strings.HasPrefix(version, "v") { // Append "v" to match release tag
		gitTag = "v" + version
	}
	return gitTag
}

