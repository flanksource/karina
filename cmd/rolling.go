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

var rollingOpts provision.RollingOptions
var RollingRestart = &cobra.Command{
	Use:   "restart",
	Short: "Rolling restart of all nodes",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.RollingRestart(getPlatform(cmd), rollingOpts); err != nil {
			log.Fatalf("Failed to restart nodes, %s", err)
		}
	},
}

var RollingUpdate = &cobra.Command{
	Use:   "update",
	Short: "Rolling update of all nodes",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.RollingUpdate(getPlatform(cmd), rollingOpts); err != nil {
			log.Fatalf("Failed to update nodes %s", err)
		}
	},
}

func init() {
	rollingOpts = provision.RollingOptions{}
	RollingUpdate.Flags().DurationVar(&rollingOpts.MinAge, "min-age", time.Hour*24*7, "Minimum age of VM's to update")
	Rolling.PersistentFlags().DurationVar(&rollingOpts.Timeout, "timeout", time.Minute*2, "timeout between actions")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.MigrateLocalVolumes, "migrate-local-volumes", true, "Delete and recreate local PVC's")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.Force, "force", false, "ignore errors and continue with the rolling action regardless of health")
	Rolling.AddCommand(RollingRestart, RollingUpdate)
}
