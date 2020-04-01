package cmd

import (
	"time"

	"github.com/moshloop/platform-cli/pkg/provision"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Rolling = &cobra.Command{
	Use: "rolling",
}

var rollingTimeout, rollingAge time.Duration
var rollingForce bool
var RollingRestart = &cobra.Command{
	Use:   "restart",
	Short: "Rolling restart of all nodes",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.RollingRestart(getPlatform(cmd), rollingTimeout, rollingForce); err != nil {
			log.Fatalf("Failed to restart nodes, %s", err)
		}
	},
}

var RollingUpdate = &cobra.Command{
	Use:   "update",
	Short: "Rolling update of all nodes",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.RollingUpdate(getPlatform(cmd), rollingAge, rollingTimeout, rollingForce); err != nil {
			log.Fatalf("Failed to update nodes %s", err)
		}
	},
}

func init() {
	RollingUpdate.Flags().DurationVar(&rollingAge, "min-age", time.Hour*24*7, "Minimum age of VM to update")
	Rolling.PersistentFlags().DurationVar(&rollingTimeout, "timeout", time.Minute*2, "timeout between actions")
	Rolling.PersistentFlags().BoolVar(&rollingForce, "force", false, "ignore errors and continue with the rolling action regardless of health")
	Rolling.AddCommand(RollingRestart, RollingUpdate)
}
