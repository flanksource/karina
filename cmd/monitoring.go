package cmd

import (
	"github.com/spf13/cobra"
)

var Monitoring = &cobra.Command{
	Use: "monitoring",

	Short: "Commands for working with monitoring stack",
}

func init() {
	Monitoring.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Compiles and builds the monitoring stack",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
}
