package cmd

import (
	"github.com/spf13/cobra"
)

var Images = &cobra.Command{
	Use:   "images",
	Short: "Commands for working with docker images",
}

func init() {
	Images.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all docker images used by the platform",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
	Images.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize all platform docker images to a local registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
}
