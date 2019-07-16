package phases

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
)

func Build(cfg types.PlatformConfig) error {

	for _, spec := range cfg.Specs {
		log.Infof("Building specs in %s", spec)
		if err := utils.Exec("gomplate --input-dir \"%s\" --output-dir build -c \".=%s\"", spec, cfg.Source); err != nil {
			return err
		}
	}

	if err := utils.Exec(".bin/jsonnet -J vendor -m build build/%s | xargs -I{} sh -c 'cat {} | .bin/gojsontoyaml > {}.yaml; rm -f {}' -- {}", "monitoring.jsonnet"); err != nil {
		return err
	}

	return nil
}
