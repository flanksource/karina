package harbor

import (
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/commons/files"
	"github.com/moshloop/commons/is"
	"github.com/moshloop/commons/text"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"os"
)

func Deploy(p *platform.Platform) error {
	dex.Defaults(p)
	defaults(p)
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
	if !is.File("build/harbor") {
		if err := files.Getter("github.com/goharbor/harbor-helm?ref=v1.1.2", "build/harbor"); err != nil {
			return err
		}
	}

	values, err := text.Template(harborYml, p.PlatformConfig)
	if err != nil {
		return err
	}
	log.Tracef("Config: \n%s\n", values)
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		return err
	}
	helm := deps.BinaryWithEnv("helm", p.Versions["helm"], ".bin", map[string]string{"KUBECONFIG": kubeconfig})
	valuesFile := text.ToFile(values, ".yml")
	defer os.Remove(valuesFile)
	helm("upgrade harbor  build/harbor -f %s --install --namespace harbor", valuesFile)

	client := NewHarborClient(p)
	return client.UpdateSettings(*p.Harbor.Settings)

}
