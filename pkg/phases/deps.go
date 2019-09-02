package phases

import (
	"github.com/moshloop/commons/deps"
	"github.com/moshloop/platform-cli/pkg/types"

	"os"
)

func Deps(cfg types.PlatformConfig) error {
	os.Mkdir(".bin", 0755)
	return deps.InstallDependencies(cfg.Versions, ".bin")
}
