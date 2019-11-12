package pgo

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/exec"
	"github.com/moshloop/commons/files"
	"github.com/moshloop/commons/is"
	"github.com/moshloop/commons/utils"
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

func getEnv(p *platform.Platform) map[string]string {
	kubeconfig, _ := p.GetKubeConfig()
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
		"PGO_CA_CERT":            "build/pgo/conf/postgres-operator/server.crt",
		PGO_CLIENT_CERT:          "build/pgo/conf/postgres-operator/server.crt",
		PGO_CLIENT_KEY:           "build/pgo/conf/postgres-operator/server.key",

		// 3.5.4 vars
		"CO_IMAGE_PREFIX":  "crunchydata",
		"CO_CMD":           "kubectl",
		"CO_UI":            "false",
		"CO_NAMESPACE":     PGO,
		"COROOT":           "build/pgo",
		"CO_IMAGE_TAG":     "centos7-" + strings.ReplaceAll(p.PGO.Version, "v", ""),
		"CO_APISERVER_URL": fmt.Sprintf("https://postgres-operator.%s", p.Domain),
		"CO_CA_CERT":       "build/pgo/conf/postgres-operator/server.crt",
		CO_CLIENT_CERT:     "build/pgo/conf/postgres-operator/server.crt",
		CO_CLIENT_KEY:      "build/pgo/conf/postgres-operator/server.key",
		PGOUSER:            os.Getenv("HOME") + "/.pgouser-" + p.Name,
	}
}

func ClientSetup(p *platform.Platform) error {
	if p.PGO == nil || p.PGO.Disabled {
		return nil
	}
	ENV := getEnv(p)

	if p.DryRun {
		return nil
	}

	user, pass := getPgoAuth(p)

	passwd := fmt.Sprintf("%s:%s", user, pass)
	home, _ := getEnv(p)["PGOUSER"]
	log.Debugf("Writing %s", home)
	if err := ioutil.WriteFile(home, []byte(passwd), 0644); err != nil {
		return err
	}

	if !is.File(ENV["PGO_CLIENT_CERT"]) {
		secrets := *p.GetSecret("pgo", "pgo-auth-secret")

		crt := home + "/.pgoserver.crt"
		key := home + "/.pgoserver.key"
		log.Debugf("Writing %s", crt)
		if err := ioutil.WriteFile(crt, secrets["server.crt"], 0644); err != nil {
			return err
		}

		log.Debugf("Writing %s", key)
		if err := ioutil.WriteFile(key, secrets["server.key"], 0644); err != nil {
			return err
		}
		ENV[PGO_CLIENT_CERT] = crt
		ENV[PGO_CA_CERT] = crt
		ENV[PGO_CLIENT_KEY] = key
		ENV[CO_CLIENT_CERT] = crt
		ENV[CO_CLIENT_KEY] = key
		ENV[CO_CA_CERT] = crt

	}
	for k, v := range ENV {
		fmt.Printf("export %s=%s\n", k, v)
	}
	deps.InstallDependency("pgo", p.PGO.Version, ".bin")
	return nil
}

func Install(p *platform.Platform) error {
	if p.PGO == nil || p.PGO.Disabled {
		return nil
	}
	ENV := getEnv(p)
	for k, v := range ENV {
		log.Tracef("export %s=%s\n", k, v)
	}

	gitTag := p.PGO.Version
	if strings.Contains(gitTag, "3.5.4") {
		gitTag = strings.ReplaceAll(gitTag, "v", "")
	}
	if err := files.Getter("git::https://github.com/CrunchyData/postgres-operator.git?ref="+gitTag, "build/pgo"); err != nil {
		return err
	}

	var passwd string
	_, pass := getPgoAuth(p)

	if pass != "" {
		log.Infof("Using existing admin password \"%s\"", pass)
		passwd = fmt.Sprintf("admin:%s:pgoadmin", pass)
	} else {
		passwd = fmt.Sprintf("admin:%s:pgoadmin", utils.RandomString(10))
	}

	pgouser := "build/pgo/conf/postgres-operator/pgouser"
	kubectl := p.GetKubectl()

	ioutil.WriteFile(pgouser, []byte(passwd), 0644)

	home, _ := getEnv(p)["PGOUSER"]
	log.Debugf("Writing %s", home)
	if err := ioutil.WriteFile(home, []byte(passwd), 0644); err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		// cp -R behavior seems to handle directories differently on macosx and linux?
		exec.ExecfWithEnv("cp -Rv overlays/pgo/ build/pgo", ENV)
	} else {
		exec.ExecfWithEnv("cp -Rv overlays/pgo/ build/", ENV)
	}
	kubectl("create ns " + PGO)

	if err := p.ExposeIngressTLS("pgo", "postgres-operator", 8443); err != nil {
		return err
	}

	if p.DryRun {
		return nil
	}
	if err := exec.ExecfWithEnv("/bin/bash  build/pgo/deploy/install-rbac.sh", ENV); err != nil {
		return err
	}
	return exec.ExecfWithEnv("/bin/bash build/pgo/deploy/deploy.sh", ENV)
}

func GetPGO(p *platform.Platform) (deps.BinaryFunc, error) {
	env := getEnv(p)
	return deps.BinaryWithEnv(PGO, p.PGO.Version, ".bin", env), nil
}
