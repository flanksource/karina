package cmd

import (
	"github.com/moshloop/platform-cli/pkg/phases"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var Dependencies = &cobra.Command{
	Use:     "dependencies",
	Aliases: []string{"deps"},
	Short:   "Installs/Updates all required dependencies for building and deploying a platform",
	Args:    cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if err := phases.Deps(getConfig(cmd)); err != nil {
			log.Fatalf("Failed installing dependencies %s", err)
		}
	},
}
