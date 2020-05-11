package cmd

import (
	"github.com/flanksource/commons/logger"
	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/spf13/cobra"
)

var Status = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the cluster and each node",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Status(getPlatform(cmd)); err != nil {
			logger.Fatalf("Failed to get cluster status, %s", err)
		}
	},
}
