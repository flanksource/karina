package cmd

import (
	"time"

	"github.com/flanksource/karina/pkg/provision"
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
	RollingUpdate.Flags().DurationVar(&rollingOpts.MinAge, "min-age", time.Hour*24*7, "Minimum age of nodes to roll")
	Rolling.PersistentFlags().DurationVar(&rollingOpts.Timeout, "timeout", time.Minute*5, "timeout between actions")
	Rolling.PersistentFlags().IntVar(&rollingOpts.Max, "max", 100, "Max number of nodes to roll")
	RollingUpdate.PersistentFlags().IntVar(&rollingOpts.MaxSurge, "max-surge", 3, "Max number of nodes surge to, the higher the number the faster the rollout, but the more capacity that will be used ")
	RollingUpdate.PersistentFlags().IntVar(&rollingOpts.HealthTolerance, "health-tolerance", 1, "Max number of failing pods to tolerate")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.MigrateLocalVolumes, "migrate-local-volumes", true, "Delete and recreate local PVC's")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.Force, "force", false, "ignore errors and continue with the rolling action regardless of health")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.Masters, "masters", true, "include master nodes")
	Rolling.PersistentFlags().BoolVar(&rollingOpts.Workers, "workers", true, "include worker nodes")
	Rolling.AddCommand(RollingRestart, RollingUpdate)
}
