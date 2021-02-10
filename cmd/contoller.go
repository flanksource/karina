package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/controller/burnin"
)

var burninControllerPeriod time.Duration
var BurninController = &cobra.Command{
	Use: "burnin-controller",
	// Short: "Commands for provisioning clusters and VMs",
	Run: func(cmd *cobra.Command, args []string) {
		// first we start a burnin controller in the background that checks
		// new nodes with the burnin taint for health, removing the taint
		// once they become healthy
		burninCancel := make(chan bool)
		burnin.Run(getPlatform(cmd), burninControllerPeriod, burninCancel)
	},
}

func init() {
	BurninController.Flags().DurationVar(&burninControllerPeriod, "burnin-period", time.Minute*3, "Period to burn-in new nodes before scheduling workloads on")
}
