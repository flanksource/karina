package phases

import (
	"os"

	"github.com/flanksource/commons/deps"
	"github.com/moshloop/platform-cli/pkg/types"
)

func Deps(cfg types.PlatformConfig) error {
	os.Mkdir(".bin", 0755) // nolint: errcheck

	return deps.InstallDependencies(cfg.Versions, ".bin")
}
