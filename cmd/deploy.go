package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	deploy_base "github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
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

	var _opa = &cobra.Command{
		Use:   "opa",
		Short: "Build and deploy opa aka gatekeeper",
	}

	_opa.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install opa control plane into the cluster",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opa.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error initializing opa %s", err)
			}
		},
	})

	_opa.AddCommand(&cobra.Command{
		Use:   "policies",
		Short: "deploy opa policies into the cluster",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opa.Deploy(getPlatform(cmd), args[0]); err != nil {
				log.Fatalf("Error deploying opa policies %s", err)
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
				log.Fatalf("Error deploying base: %s", err)
			}
			if err := pgo.Install(p); err != nil {
				log.Warnf("Error deployed postgres operator: %v", err)
			}
			if err := pgo.ClientSetup(p); err != nil {
				log.Warnf("Error deployed pgo client: %v", err)
			}
			if err := monitoring.Install(p); err != nil {
				log.Warnf("Error building monitoring stack: %v", err)
			}
			if err := harbor.Deploy(p); err != nil {
				log.Warnf("Error deploying harbor: %v", err)
			}
			if err := dex.Install(p); err != nil {
				log.Warnf("Error initializing dex: %v", err)
			}
			if err := opa.Install(p); err != nil {
				log.Fatalf("Error installing opa control plane: %s", err)
			}
			if err := flux.Install(p); err != nil {
				log.Fatalf("Error installing flux: %s", err)
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
		Use:   "nsx",
		Short: "Build and deploy the NSX-T CNI",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := nsx.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying nsx %s", err)
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

	Deploy.AddCommand(&cobra.Command{
		Use:   "stubs",
		Short: "Build and deploy stubs for integration testing",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := stubs.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploy stubs %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "gitops",
		Short: "Build and deploy Flux gitops agents",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := flux.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploy flux %s", err)
			}
		},
	})

	Deploy.AddCommand(_pgo, _opa, all)
}
