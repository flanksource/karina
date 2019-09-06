package cmd

import (
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"log"

	"github.com/spf13/cobra"
)

var Deploy = &cobra.Command{
	Use:   "deploy",
	Short: "Build the platform",
}

func init() {
	Deploy.PersistentFlags().Bool("dry-run", false, "Don't execute anything")

	var _pgo = &cobra.Command{
		Use:   "pgo",
		Short: "Build and deploy the postgres operator",
	}

	_pgo.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install the PostgreOperator into the cluster",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := pgo.Install(getPlatform(cmd), dryRun); err != nil {
				log.Fatalf("Error deployed postgres operator%s", err)
			}
		},
	})

	_pgo.AddCommand(&cobra.Command{
		Use:   "client",
		Short: "Setup the the pgo client",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := pgo.ClientSetup(getPlatform(cmd), dryRun); err != nil {
				log.Fatalf("Error deployed pgo client operator%s", err)
			}
		},
	})

	var all = &cobra.Command{
		Use:   "all",
		Short: "Build everything",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := pgo.Install(getPlatform(cmd), dryRun); err != nil {
				log.Fatalf("Error deployed postgres operator%s", err)
			}
			if err := pgo.ClientSetup(getPlatform(cmd), dryRun); err != nil {
				log.Fatalf("Error deployed pgo client operator%s", err)
			}
			if err := monitoring.Install(getConfig(cmd)); err != nil {
				log.Fatalf("Error building monitoring stack %s", err)
			}
			if err := dex.Install(getConfig(cmd)); err != nil {
				log.Fatalf("Error initializing dex %s", err)
			}
		},
	}
	Deploy.AddCommand(&cobra.Command{
		Use:   "monitoring",
		Short: "Build and deploy the prometheus/grafana monitoring stack",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := monitoring.Install(getConfig(cmd)); err != nil {
				log.Fatalf("Error building monitoring stack %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "harbor",
		Short: "Build and deploy the harbor registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := harbor.Install(getPlatform(cmd), dryRun); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "dex",
		Short: "Build and deploy the dex-ca",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := dex.Install(getConfig(cmd)); err != nil {
				log.Fatalf("Error initializing dex %s", err)
			}
		},
	})

	Deploy.AddCommand(_pgo, all)
}
