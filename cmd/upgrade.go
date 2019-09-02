package cmd

import (
	"github.com/spf13/cobra"
)

var Upgrade = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the core platform components to their latest versions",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
	},
}
