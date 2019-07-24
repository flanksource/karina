package cmd

import (
	"io/ioutil"
	"log"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func getConfig(cmd *cobra.Command) types.PlatformConfig {
	path, _ := cmd.Flags().GetString("config")
	cfg := types.PlatformConfig{
		Source: path,
	}
	if path == "" {
		return cfg
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file %s, %s", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse YAML: %s", err)
	}

	monitoring, _ := cmd.Flags().GetBool("monitoring")
	cfg.BuildOptions.Monitoring = monitoring
	cfg.Init()
	return cfg
}
