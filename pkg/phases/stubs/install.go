package stubs

import "github.com/moshloop/platform-cli/pkg/platform"

func Install(platform *platform.Platform) error {
	return platform.ApplySpecs("", "stubs/")
}
