package cmd

import (
	"github.com/flanksource/commons/logger"
	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/spf13/cobra"
)

var Upgrade = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the kubernetes control plane",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Upgrade(getPlatform(cmd)); err != nil {
			logger.Fatalf("Failed to upgrade cluster, %s", err)
		}
	},
}
