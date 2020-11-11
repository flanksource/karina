package cmd

import (
	"time"

	"github.com/flanksource/karina/pkg/operator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Operator = &cobra.Command{
	Use:   "operator",
	Short: "Run karina operator",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		metricsAddr, _ := cmd.Flags().GetString("metrics-addr")
		enableLeaderElection, _ := cmd.Flags().GetBool("enable-leader-election")
		syncPeriod, _ := cmd.Flags().GetDuration("sync-period")
		logLevel, _ := cmd.Flags().GetString("log-level")
		port, _ := cmd.Flags().GetInt("port")

		operatorConfig := operator.Config{
			MetricsAddr:          metricsAddr,
			EnableLeaderElection: enableLeaderElection,
			SyncPeriod:           syncPeriod,
			LogLevel:             logLevel,
			Port:                 port,
		}

		op, err := operator.New(operatorConfig)
		if err != nil {
			log.Fatalf("failed to create operator: %v", err)
		}

		if err := op.Run(); err != nil {
			log.Fatalf("failed to start operator: %v", err)
		}
	},
}

func init() {
	Operator.Flags().String("metrics-addr", ":8080", "The address the metrics endpoint binds to.")
	Operator.Flags().Bool("enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	Operator.Flags().Duration("sync-period", 300*time.Second, "The resync period used for reconciling")
	Operator.Flags().String("log-level", "error", "Logging level: debug, info, error")
	Operator.Flags().Int("port", 9443, "Port to run the web server")
}
