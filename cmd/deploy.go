package cmd

import (
	"os"

	log "github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/phases/base"
	"github.com/flanksource/karina/pkg/phases/calico"
	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/phases/configmapreloader"
	"github.com/flanksource/karina/pkg/phases/dex"
	"github.com/flanksource/karina/pkg/phases/eck"
	"github.com/flanksource/karina/pkg/phases/elasticsearch"
	"github.com/flanksource/karina/pkg/phases/filebeat"
	"github.com/flanksource/karina/pkg/phases/fluentdoperator"
	"github.com/flanksource/karina/pkg/phases/flux"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/phases/monitoring"
	"github.com/flanksource/karina/pkg/phases/nsx"
	"github.com/flanksource/karina/pkg/phases/opa"
	"github.com/flanksource/karina/pkg/phases/platformoperator"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/phases/registrycreds"
	"github.com/flanksource/karina/pkg/phases/sealedsecrets"
	"github.com/flanksource/karina/pkg/phases/stubs"
	"github.com/flanksource/karina/pkg/phases/vault"
	"github.com/flanksource/karina/pkg/phases/velero"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/spf13/cobra"
)

type DeployFn func(p *platform.Platform) error

var Phases = map[string]DeployFn{
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

var PhasesExtra = map[string]DeployFn{
	"cert-manager":      certmanager.Install,
	"platform-operator": platformoperator.Install,
	"vsphere":           vsphere.Install,
}

var PhaseOrder = []string{"calico", "nsx", "base", "stubs", "postgres-operator", "dex", "vault"}

var Deploy = &cobra.Command{
	Use: "deploy",
}

func init() {
	var PhasesCmd = &cobra.Command{
		Use: "phases",
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			// we track the failure status, and continue on failure to allow degraded operations
			failed := false
			// first deploy strictly ordered phases, these phases are often dependencies for other phases
			for _, name := range PhaseOrder {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				if err := Phases[name](p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, err)
					failed = true
				}
				// remove the phase from the map so it isn't run again
				delete(Phases, name)
			}
			for name, fn := range Phases {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				if err := fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, err)
					failed = true
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}

	Deploy.AddCommand(PhasesCmd)

	for name, fn := range Phases {
		_name := name
		_fn := fn
		PhasesCmd.Flags().Bool(name, false, "Deploy "+name)
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

	for name, fn := range PhasesExtra {
		_name := name
		_fn := fn
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
			p := getPlatform(cmd)
			// we track the failure status, and continue on failure to allow degraded operations
			failed := false

			// first deploy strictly ordered phases, these phases are often dependencies for other phases
			for _, name := range PhaseOrder {
				if err := Phases[name](p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, err)
					failed = true
				}
				// remove the phase from the map so it isn't run again
				delete(Phases, name)
			}

			for name, fn := range Phases {
				if err := fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, err)
					failed = true
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}

	Deploy.AddCommand(all)
}
