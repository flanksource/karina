package cmd

import (
	"log"

	"github.com/moshloop/platform-cli/pkg/phases"

	"github.com/spf13/cobra"
)

var Build = &cobra.Command{
	Use:   "build",
	Short: "Build the platform",
}

var dex = &cobra.Command{
	Use:   "dex",
	Short: "Build the dex-ca",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Dex(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing dex %s", err)
		}
	},
}

var monitoring = &cobra.Command{
	Use:   "monitoring",
	Short: "Build the prometheus/grafana monitoring stack",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Monitoring(getConfig(cmd)); err != nil {
			log.Fatalf("Error building monitoring stack %s", err)
		}
	},
}

var base = &cobra.Command{
	Use:   "base",
	Short: "Build the base platform configuration from ",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Build(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing repo %s", err)
		}

	},
}

var all = &cobra.Command{
	Use:   "all",
	Short: "Build everything",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Build(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing repo %s", err)
		}
		if err := phases.Dex(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing dex %s", err)
		}
		if err := phases.Monitoring(getConfig(cmd)); err != nil {
			log.Fatalf("Error building monitoring stack %s", err)
		}
	},
}

func init() {
	Build.AddCommand(dex, monitoring, base, all)
}
