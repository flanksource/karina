package cmd

import (
	"github.com/moshloop/platform-cli/pkg/phases/elasticsearch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	deploy_base "github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/certmanager"
	"github.com/moshloop/platform-cli/pkg/phases/configmapreloader"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/eck"
	"github.com/moshloop/platform-cli/pkg/phases/filebeat"
	"github.com/moshloop/platform-cli/pkg/phases/fluentdoperator"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nginx"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/postgresoperator"
	"github.com/moshloop/platform-cli/pkg/phases/registrycreds"
	"github.com/moshloop/platform-cli/pkg/phases/sealedsecrets"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
	"github.com/moshloop/platform-cli/pkg/phases/vault"
	"github.com/moshloop/platform-cli/pkg/phases/velero"
)

var Deploy = &cobra.Command{
	Use:   "deploy",
	Short: "Build the platform",
}

func init() {
	var _opa = &cobra.Command{
		Use:   "opa",
		Short: "Build and deploy opa aka gatekeeper",
	}

	_opa.AddCommand(&cobra.Command{
		Use:   "bundle",
		Short: "deploy opa bundle",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opa.DeployBundle(getPlatform(cmd), args[0]); err != nil {
				log.Fatalf("Error deploying  opa bundles: %s", err)
			}
		},
	})

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
			if err := postgresoperator.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying postgres-operator %s", err)
			}
			if err := eck.Deploy(p); err != nil {
				log.Fatalf("Error deploying ECK: %s", err)
			}
			if err := monitoring.Install(p); err != nil {
				log.Warnf("Error deploying monitoring stack: %v", err)
			}
			if err := harbor.Deploy(p); err != nil {
				log.Warnf("Error deploying harbor: %v", err)
			}
			if err := dex.Install(p); err != nil {
				log.Warnf("Error deploying dex: %v", err)
			}
			if err := opa.Install(p); err != nil {
				log.Fatalf("Error deploying opa control plane: %s", err)
			}
			if err := flux.Install(p); err != nil {
				log.Fatalf("Error deploying flux: %s", err)
			}
			if err := velero.Install(p); err != nil {
				log.Fatalf("Error deploying velero: %s", err)
			}
			if err := fluentdoperator.Deploy(p); err != nil {
				log.Fatalf("Error deploying fluentd: %s", err)
			}
			if err := filebeat.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying filebeat %s", err)
			}
			if err := sealedsecrets.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying sealed secrets %s", err)
			}
			if err := vault.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying vault %s", err)
			}
			if err := configmapreloader.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying configmap-reloader %s\n", err)
			}
			if err := registrycreds.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying registry credentials %s\n", err)
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
				log.Fatalf("Error deploying dex %s", err)
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
		Use:   "certmanager",
		Short: "Build and deploy the certmanager",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := certmanager.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying cert manager %s", err)
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
		Use:   "velero",
		Short: "Deploy velero for backups",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := velero.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying velero %s", err)
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

	Deploy.AddCommand(&cobra.Command{
		Use:   "fluentd",
		Short: "Deploy the fluentd operator",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := fluentdoperator.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying fluentd operator %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "eck",
		Short: "Deploy the eck operator",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := eck.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying eck operator %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "postgres-operator",
		Short: "Deploy the zalando postgres-operator",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := postgresoperator.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying postgres-operator %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "filebeat",
		Short: "Deploy filebeat",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := filebeat.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying filebeat %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "nginx",
		Short: "Deploy nginx",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := nginx.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying nginx %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "configmap-reloader",
		Short: "Deploy configmap-reloader",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := configmapreloader.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying configmap-reloader %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "sealed-secrets",
		Short: "Deploy sealed secrets controller",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := sealedsecrets.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying sealed secrets controller %s\n", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "vault",
		Short: "Deploy vault",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := vault.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying vault %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "elasticsearch",
		Short: "Deploy Elasticsearch",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := elasticsearch.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying elasticsearch %s", err)
			}
		},
	})

	Deploy.AddCommand(&cobra.Command{
		Use:   "registry-credentials",
		Short: "Deploy registry credentials controller",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := registrycreds.Install(getPlatform(cmd)); err != nil {
				log.Fatalf("Error deploying registry credentials controller %s\n", err)
			}
		},
	})

	Deploy.AddCommand(_opa, all)
}
