package cmd

import (
	"log"

	"github.com/spf13/cobra"

	deploy_base "github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
)

var Deploy = &cobra.Command{
	Use:   "deploy",
	Short: "Build the platform",
}

func init() {

	var _pgo = &cobra.Command{
		Use:   "pgo",
		Short: "Build and deploy the postgres operator",
	}

	_pgo.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install the PostgreOperator into the cluster",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := pgo.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deployed postgres operator%s", err)
			}
		},
	})

	_pgo.AddCommand(&cobra.Command{
		Use:   "client",
		Short: "Setup the the pgo client",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := pgo.ClientSetup(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deployed pgo client operator%s", err)
			}
		},
	})

	var all = &cobra.Command{
		Use:   "all",
		Short: "Build everything",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			if err := deploy_base.Install(p); err != nil {
				log.Fatalf("Error deploying base %s", err)
			}
			if err := pgo.Install(p); err != nil {
				log.Fatalf("Error deployed postgres operator%s", err)
			}
			if err := pgo.ClientSetup(p); err != nil {
				log.Fatalf("Error deployed pgo client operator%s", err)
			}
			if err := monitoring.Install(p); err != nil {
				log.Fatalf("Error building monitoring stack %s", err)
			}
			if err := harbor.Deploy(p); err != nil {
				log.Fatalf("Error deploying harbor %s", err)
			}
			if err := dex.Install(p); err != nil {
				log.Fatalf("Error initializing dex %s", err)
			}
		},
	}
	Deploy.AddCommand(&cobra.Command{
		Use:   "monitoring",
		Short: "Build and deploy the prometheus/grafana monitoring stack",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := monitoring.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building monitoring stack %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "dex",
		Short: "Build and deploy the dex-ca",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := dex.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error initializing dex %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "calico",
		Short: "Build and deploy calico",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := calico.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploy calico dex %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "harbor",
		Short: "Build and deploy the harbor registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "base",
		Short: "Build and deploy base dependencies",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := deploy_base.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploy base %s", err)
			}
		},
	})
	Deploy.AddCommand(_pgo, all)
}
