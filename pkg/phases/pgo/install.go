package pgo

import (
	"fmt"
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/exec"
	"github.com/moshloop/commons/files"
	"github.com/moshloop/commons/is"
	"github.com/moshloop/commons/utils"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	PGO     = "pgo"
	PGOUSER = "PGOUSER"
	PGOPASS = "PGOPASS"
)

func getPgoAuth(p *platform.Platform) (user, pass string) {
	secret := p.GetSecret("pgo", "pgo-auth-secret")
	if secret != nil {
		pgouser := string((*secret)["pgouser"])
		user = strings.Split(pgouser, ":")[0]
		pass = strings.Split(pgouser, ":")[1]
	}
	return
}

func getEnv(p *platform.Platform) map[string]string {

	return map[string]string{
		"PATH":       ".bin:" + os.Getenv("PATH"),
		"KUBECONFIG": p.Name + "-admin.yml",

		// 4.0.0 vars
		"PGO_OPERATOR_NAMESPACE": PGO,
		"NAMESPACE":              "pgo-databases",
		"PGO_APISERVER_URL":      fmt.Sprintf("https://pgo.%s", p.Domain),
		"PGO_CMD":                ".bin/kubectl",
		"PGOROOT":                "build/pgo",
		"PGO_IMAGE_PREFIX":       "crunchydata",
		"PGO_BASEOS":             "centos7",
		"PGO_VERSION":            "3.5.4",
		"PGO_IMAGE_TAG":          "centos7-4.0.1",
		"PGO_CA_CERT":            "build/pgo/conf/postgres-operator/server.crt",
		"PGO_CLIENT_CERT":        "build/pgo/conf/postgres-operator/server.crt",
		"PGO_CLIENT_KEY":         "build/pgo/conf/postgres-operator/server.key",

		// 3.5.4 vars
		"CO_IMAGE_PREFIX":  "crunchydata",
		"CO_CMD":           "kubectl",
		"CO_NAMESPACE":     PGO,
		"COROOT":           "build/pgo",
		"CO_IMAGE_TAG":     "centos7-3.5.4",
		"CO_APISERVER_URL": fmt.Sprintf("https://pgo.%s", p.Domain),
		"CO_CA_CERT":       "build/pgo/conf/postgres-operator/server.crt",
		"CO_CLIENT_CERT":   "build/pgo/conf/postgres-operator/server.crt",
		"CO_CLIENT_KEY":    "build/pgo/conf/postgres-operator/server.key",
		PGOUSER:            os.Getenv("HOME") + "/.pgouser",
	}
}

func ClientSetup(p *platform.Platform, dryRun bool) error {
	ENV := getEnv(p)
	for k, v := range ENV {
		fmt.Printf("export %s=%s\n", k, v)
	}
	if dryRun {
		return nil
	}

	user, pass := getPgoAuth(p)

	passwd := fmt.Sprintf("%s:%s:pgoadmin", user, pass)
	home, _ := os.UserHomeDir()
	log.Debugf("Writing %s/.pgouser", home)
	if err := ioutil.WriteFile(home+"/.pgouser", []byte(passwd), 0644); err != nil {
		return err
	}

	return nil
}

func Install(p *platform.Platform, dryRun bool) error {
	ENV := getEnv(p)
	for k, v := range ENV {
		log.Tracef("export %s=%s\n", k, v)
	}
	if dryRun {
		return nil
	}

	if !is.File("build/pgo") {
		if err := files.Getter("git::https://github.com/CrunchyData/postgres-operator.git?ref="+p.PGO.Version, "build/pgo"); err != nil {
			return err
		}
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
	kubectl := deps.Binary("kubectl", "", ".bin")

	ioutil.WriteFile(pgouser, []byte(passwd), 0644)
	exec.ExecfWithEnv("cp -R overlays/pgo/ $PGOROOT", ENV)
	kubectl("create ns " + PGO)
	kubectl("create ns pgo-databases")
	if err := exec.ExecfWithEnv("/bin/bash  build/pgo/deploy/install-rbac.sh", ENV); err != nil {
		return err
	}
	return exec.ExecfWithEnv("/bin/bash build/pgo/deploy/deploy.sh", ENV)
}

func GetPGO(p *platform.Platform) (deps.BinaryFunc, error) {
	env := getEnv(p)
	return deps.BinaryWithEnv(PGO, p.PGO.Version, ".bin", env), nil
}

// GetOrCreateDB creates a new postgres cluster and returns access details
func GetOrCreateDB(p *platform.Platform, name string, replicas int, databases ...string) (*types.DB, error) {
	pgo, err := GetPGO(p)
	if err != nil {
		return nil, err
	}

	secret := p.GetSecret(PGO, name+"-postgres-secret")
	var passwd string
	if secret != nil {
		log.Infof("Reusing database %s\n", name)
		passwd = string((*secret)["password"])
	} else {
		log.Infof("Creating new database %s\n", name)
		passwd = utils.RandomString(10)
		if err := pgo("create cluster %s -w %s --replica-count %d --debug", name, passwd, replicas); err != nil {
			return nil, err
		}
	}

	return &types.DB{
		Host:     fmt.Sprintf("%s.%s.svc.cluster.local", name, PGO),
		Username: "postgres",
		Password: passwd,
		Port:     5432,
	}, nil
}

func WaitForDB(p *platform.Platform, db string) error {
	kubectl := p.GetKubectl()
	for {
		if err := kubectl(" -n pgo exec svc/%s bash -c database -- -c \"psql -c 'SELECT 1';\"", db); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func CreateDatabase(p *platform.Platform, db string, databases ...string) error {
	kubectl := p.GetKubectl()
	for _, database := range databases {
		if err := kubectl("-n pgo exec svc/%s bash -c database -- -c \"psql -c 'create database %s'\"", db, database); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}
