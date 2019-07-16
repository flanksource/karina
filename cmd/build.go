package cmd

import (
	"log"

	"github.com/moshloop/platform-cli/pkg/phases"

	"github.com/spf13/cobra"
)

var Build = &cobra.Command{
	Use:   "build",
	Short: "Build the platform",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if err := phases.Build(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing repo %s", err)
		}

	},
}
