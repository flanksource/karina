package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/spf13/cobra"
)

var Provision = &cobra.Command{
	Use:   "provision",
	Short: "Provision a new cluster",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if err := provision.Provision(getConfig(cmd)); err != nil {
			log.Fatalf("Failed to provision cluster, %s", err)
		}
	},
}
