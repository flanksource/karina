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

var helm = &cobra.Command{
	Use:   "helm",
	Short: "Build and template helm charts",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Helm(getConfig(cmd)); err != nil {
			log.Fatalf("Error building helm templates %s", err)
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

		if err := phases.Helm(getConfig(cmd)); err != nil {
			log.Fatalf("Error building helm templates %s", err)
		}
	},
}

func init() {
	Build.AddCommand(helm, base, all)
}
