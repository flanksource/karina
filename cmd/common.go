package cmd

import (
	"github.com/imdario/mergo"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func getPlatform(cmd *cobra.Command) *platform.Platform {
	platform := platform.Platform{
		PlatformConfig: getConfig(cmd),
	}
	return &platform
}

func getConfig(cmd *cobra.Command) types.PlatformConfig {
	paths, _ := cmd.Flags().GetStringArray("config")
	base := types.PlatformConfig{
		Source: paths[0],
	}

	for _, path := range paths {
		cfg := types.PlatformConfig{
			Source: paths[0],
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read config file %s, %s", path, err)
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("Failed to parse YAML: %s", err)
		}

		for node, vm := range cfg.Nodes {
			if baseNode, ok := base.Nodes[node]; ok {
				if err := mergo.Merge(&baseNode, vm); err != nil {
					log.Fatalf("Failed to merge nodes %s, %s", node, err)
				}
				base.Nodes[node] = baseNode
			}
		}

		if err := mergo.Merge(&base, cfg); err != nil {
			log.Fatalf("Failed to merge in %s, %s", path, err)
		}
	}

	base.S3.AccessKey = template(base.S3.AccessKey)
	base.S3.SecretKey = template(base.S3.SecretKey)

	data, _ := yaml.Marshal(base)
	log.Tracef("Using configuration: \n%s\n", string(data))
	base.Init()
	return base
}

func template(val string) string {
	if strings.HasPrefix(val, "$") {
		env := os.Getenv(val[1:])
		if env != "" {
			return env
		}
	}
	return val
}
