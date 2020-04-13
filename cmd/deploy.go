package cmd

import (
	log "github.com/flanksource/commons/logger"
	"github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/elasticsearch"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/configmapreloader"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/eck"
	"github.com/moshloop/platform-cli/pkg/phases/filebeat"
	"github.com/moshloop/platform-cli/pkg/phases/fluentdoperator"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
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
	Use: "deploy",
}

func init() {
	type DeployFn func(p *platform.Platform) error
	phases := map[string]DeployFn{
		"base":               base.Install,
		"calico":             calico.Install,
		"configmap-reloader": configmapreloader.Deploy,
		"dex":                dex.Install,
		"eck":                eck.Deploy,
		"elasticsearch":      elasticsearch.Deploy,
		"fluentd":            fluentdoperator.Deploy,
		"filebeat":           filebeat.Deploy,
		"gitops":             flux.Install,
		"harbor":             harbor.Deploy,
		"monitoring":         monitoring.Install,
		"opa":                opa.Install,
		"nsx":                nsx.Install,
		"postgres-operator":  postgresoperator.Deploy,
		"registry-creds":     registrycreds.Install,
		"sealed-secrets":     sealedsecrets.Install,
		"stubs":              stubs.Install,
		"vault":              vault.Deploy,
		"velero":             velero.Install,
	}

	order := []string{"calico", "nsx", "base", "stubs", "postgres-operator", "dex", "vault"}

	// pinpoint := map[string]DeployFn{
	// 	"cert-manager": certmanager.Install,
	// 	"quack":        quack.Install,
	// }

	var Phases = &cobra.Command{
		Use: "phases",
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			// first deploy strictly ordered phases, these phases are often dependencies for other phases
			for _, name := range order {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				if err := phases[name](p); err != nil {
					log.Fatalf("Failed to deploy %s: %v", name, err)
				}
				// remove the phase from the map so it isn't run again
				delete(phases, name)
			}
			for name, fn := range phases {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				if err := fn(p); err != nil {
					log.Fatalf("Failed to deploy %s: %v", name, err)
				}
			}
		},
	}

	Deploy.AddCommand(Phases)

	for name, fn := range phases {
		_name := name
		_fn := fn
		Phases.Flags().Bool(name, false, "Deploy "+name)
		Deploy.AddCommand(&cobra.Command{
			Use:  name,
			Args: cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				p := getPlatform(cmd)
				if err := _fn(p); err != nil {
					log.Fatalf("Failed to deploy %s: %v", _name, err)
				}
			},
		})
	}

	var all = &cobra.Command{
		Use:   "all",
		Short: "Build everything",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			for name, fn := range phases {
				if err := fn(p); err != nil {
					log.Fatalf("Failed to deploy %s: %v", name, err)
				}
			}
		},
	}

	Deploy.AddCommand(all)
}
