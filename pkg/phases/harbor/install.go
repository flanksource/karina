package harbor

import (
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/text"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"os"
)

func Install(p *platform.Platform, dryRun bool) error {
	if p.Harbor.DB == nil {
		db, err := pgo.GetOrCreateDB(p, "harbor", 3)
		if err != nil {
			return err
		}
		if err := pgo.WaitForDB(p, "harbor"); err != nil {
			return err
		}

		if err := pgo.CreateDatabase(p, "harbor", "registry", "clair", "notary_server", "notary_signer"); err != nil {
			return err
		}
		p.Harbor.DB = db
	}

	values, err := text.Template(harborYml, p.PlatformConfig)
	if err != nil {
		return err
	}
	log.Infof("Config: \n%s\n", values)
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		return err
	}
	helm := deps.BinaryWithEnv("helm", p.Versions["helm"], ".bin", map[string]string{"KUBECONFIG": kubeconfig})
	valuesFile := text.ToFile(values, ".yml")
	defer os.Remove(valuesFile)
	return helm("install build/harbor -f %s -n harbor --namespace harbor", valuesFile)
}
