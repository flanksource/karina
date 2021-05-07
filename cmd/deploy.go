package cmd

import (
	"os"

	log "github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/phases/order"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var deployExclude []string
var Deploy = &cobra.Command{
	Use: "deploy",
}

func init() {
	var PhasesCmd = &cobra.Command{
		Use: "phases",
		Run: func(cmd *cobra.Command, args []string) {
			phases := order.GetAllPhases()
			p := getPlatform(cmd)
			if _, err := p.GetClientset(); err != nil {
				log.Fatalf("Failed to connect to platform, aborting deployment: %s", err)
				os.Exit(1)
			}
			// we track the failure status, and continue on failure to allow degraded operations
			failed := false
			// first deploy strictly ordered phases, these phases are often dependencies for other phases
			for _, name := range order.PhaseOrder {
				flag, _ := cmd.Flags().GetBool(string(name))
				if !flag {
					continue
				}
				if err := phases[name].Fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, errors.WithStack(err))
					failed = true
				}
				// remove the phase from the map so it isn't run again
				delete(phases, name)
			}
			for name, deployMap := range phases {
				flag, _ := cmd.Flags().GetBool(string(name))
				if !flag {
					continue
				}

				if err := deployMap.Fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, errors.WithStack(err))
					failed = true
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}

	Deploy.AddCommand(PhasesCmd)

	for name, deployMap := range order.GetAllPhases() {
		PhasesCmd.Flags().Bool(string(name), false, "Deploy "+string(name))
		Deploy.AddCommand(&cobra.Command{
			Use:  string(name),
			Args: cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				p := getPlatform(cmd)
				if err := deployMap.Fn(p); err != nil {
					log.Fatalf("Failed to deploy %s: %v", name, err)
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
			phases := order.GetPhases()
			all := order.GetAllPhases()
			// we track the failure status, and continue on failure to allow degraded operations
			failed := false

			for _, phase := range order.BootstrapPhases {
				if sliceContains(deployExclude, string(phase)) {
					p.Tracef("Skipping excluded phase %s", phase)
					continue
				}
				p.Tracef("Deploying %s", phase)
				if err := all[phase].Fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", phase, err)
					failed = true
				}
			}

			for name, deployMap := range phases {
				p.Tracef("Deploying %s", name)
				if err := deployMap.Fn(p); err != nil {
					log.Errorf("Failed to deploy %s: %v", name, err)
					failed = true
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}

	all.Flags().StringSliceVar(&deployExclude, "exclude", []string{}, "A list of phases to exclude from deployment")
	Deploy.AddCommand(all)
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
