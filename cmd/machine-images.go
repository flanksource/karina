package cmd

import (
	"github.com/spf13/cobra"
)

var MachineImages = &cobra.Command{
	Use:     "machine-images",
	Aliases: []string{"vm"},
	Short:   "Commands for working with machine images",
}

func init() {
	MachineImages.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all machine images currently uplooded",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
	MachineImages.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Builds a new machine-image",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
	MachineImages.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Uploads a new machine image",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
}
