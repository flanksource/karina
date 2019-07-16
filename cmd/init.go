package cmd

import (
	"log"

	"github.com/moshloop/platform-cli/pkg/phases"

	"github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new platform repository",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if err := phases.Init(getConfig(cmd)); err != nil {
			log.Fatalf("Error initializing repo %s", err)
		}

	},
}
