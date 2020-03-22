package phases

import (
	"os"

	"github.com/flanksource/commons/deps"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/pkg/errors"
)

func Deps(cfg types.PlatformConfig) error {
	if err := os.Mkdir(".bin", 0755); err != nil {
		return errors.Wrap(err, "failed to create directory .bin")
	}
	return deps.InstallDependencies(cfg.Versions, ".bin")
}
