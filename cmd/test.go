package cmd

import (
	"github.com/spf13/cobra"
)

var Test = &cobra.Command{
	Use:   "test",
	Short: "Run E2E conformance tests",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
	},
}
