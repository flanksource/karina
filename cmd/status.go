package cmd

import (
	"github.com/moshloop/platform-cli/pkg/phases"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Status = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the cluster and each node",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.Status(getPlatform(cmd)); err != nil {
			log.Fatalf("Failed to get cluster status, %s", err)
		}
	},
}
