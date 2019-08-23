package cmd

import (
	"github.com/imdario/mergo"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func getPlatform(cmd *cobra.Command) *platform.Platform {
	platform := platform.Platform{
		PlatformConfig: getConfig(cmd),
	}
	if err := platform.OpenViaEnv(); err != nil {
		log.Fatalf("Failed to initialize platform: %s", err)
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

	data, _ := yaml.Marshal(base)
	log.Debugf("Using configuration: \n%s\n", string(data))
	base.Init()
	return base
}
