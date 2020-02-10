package pgo

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pgoapi "github.com/moshloop/platform-cli/pkg/api/pgo"
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
	Namespace       = "pgo"
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
		return nil, err
	}

	secrets := *p.GetSecret("pgo", "pgo.tls")

	log.Debugf("Writing %s", ENV["PGO_CLIENT_CERT"])
	if err := ioutil.WriteFile(ENV["PGO_CLIENT_CERT"], secrets["tls.crt"], 0644); err != nil {
		return nil, err
	}

	log.Debugf("Writing %s", ENV["PGO_CLIENT_KEY"])
	if err := ioutil.WriteFile(ENV["PGO_CLIENT_KEY"], secrets["tls.key"], 0644); err != nil {
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

	if err := p.CreateOrUpdateNamespace(PGO, map[string]string{
		pgoapi.LABEL_VENDOR:                pgoapi.LABEL_CRUNCHY,
		pgoapi.LABEL_PGO_INSTALLATION_NAME: "pgo",
	}, nil); err != nil {
		return err
	}

	if p.PGO.Password == "" {
		_, pass := getPgoAuth(p)
		if pass == "" {
			pass = utils.RandomString(10)
		}
		p.PGO.Password = pass
	}
	if err := p.ApplySpecs(Namespace, "pgo-crd.yml"); err != nil {
		return err
	}
	if err := p.ApplySpecs(Namespace, "pgo-config.yml.raw"); err != nil {
		return err
	}
	if err := p.ApplySpecs(Namespace, "pgo.yml"); err != nil {
		return err
	}
	if err := p.Apply(Namespace, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pgo-backrest-repo-config",
			Labels: map[string]string{
				pgoapi.LABEL_VENDOR: pgoapi.LABEL_CRUNCHY,
			},
		},
		Data: map[string][]byte{
			"aws-s3-ca.crt":           []byte{},
			"aws-s3-credentials.yaml": []byte(fmt.Sprintf("aws-s3-key: %s\n aws-s3-key-secret: %s\n", p.S3.AccessKey, p.S3.SecretKey)),
			"config":                  []byte(pgoapi.DEFAULT_SSH_CONFIG),
			"sshd_config":             []byte(pgoapi.DEFAULT_SSHD_CONFIG),
		},
	}); err != nil {
		return err
	}

	return p.ExposeIngressTLS("pgo", "postgres-operator", 8443)
}

func GetPGO(p *platform.Platform) (deps.BinaryFunc, error) {
	env, err := getEnv(p)
	if err != nil {
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
