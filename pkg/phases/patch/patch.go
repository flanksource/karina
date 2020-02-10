package patch

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Install(platform *platform.Platform, patchYamlPath string) error {
	return platform.ApplySpecs("", patchYamlPath)
}
