package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/spf13/cobra"
)

var Cleanup = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup a cluster and terminate all VM's and resources associated with it",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Cleanup(getConfig(cmd)); err != nil {
			log.Fatalf("Failed to cleanup cluster, %s", err)
		}
	},
}
