package cmd

import (
	"time"

	"github.com/flanksource/karina/pkg/status"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/provision"
	"github.com/spf13/cobra"
)

var Status = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the cluster and each node",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Status(getPlatform(cmd)); err != nil {
			logger.Fatalf("Failed to get cluster status, %s", err)
		}
	},
}

func init() {
	var restartLimit time.Duration
	pods := &cobra.Command{
		Use:   "pods",
		Short: "Print the status of unhealthy pods across the cluster",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := provision.PodStatus(getPlatform(cmd), restartLimit); err != nil {
				logger.Fatalf("Failed to get pod status, %s", err)
			}
		},
	}

	pods.Flags().DurationVar(&restartLimit, "restart-limit", 3*time.Minute, "The previous time window in which if a pod restart it is considered unhealthy")

	Status.AddCommand(pods)

	Status.AddCommand(&cobra.Command{
		Use: "violations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return status.Violations(getPlatform(cmd))
		},
	})
}
