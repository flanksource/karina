package main

import (
	"fmt"
	"os"

	"github.com/flanksource/commons/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/moshloop/platform-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var root = &cobra.Command{
		Use:              "karina",
		PersistentPreRun: cmd.GlobalPreRun,
	}

	root.AddCommand(
		cmd.Access,
		cmd.APIDocs,
		cmd.Apply,
		cmd.Backup,
		cmd.CA,
		cmd.Cleanup,
		cmd.Config,
		cmd.Conformance,
		cmd.Consul,
		cmd.DB,
		cmd.Deploy,
		cmd.DNS,
		cmd.Exec,
		cmd.ExecNode,
		cmd.Harbor,
		cmd.Images,
		cmd.Logs,
		cmd.MachineImages,
		cmd.NSX,
		cmd.Opa,
		cmd.Provision,
		cmd.Render,
		cmd.Report,
		cmd.Rolling,
		cmd.Snapshot,
		cmd.Status,
		cmd.Test,
		cmd.TerminateNodes,
		cmd.TerminateOrphans,
		cmd.Upgrade,
		cmd.Vault,
	)

	if len(commit) > 8 {
		version = fmt.Sprintf("%v, commit %v, built at %v", version, commit[0:8], date)
	}
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of karina",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	docs := &cobra.Command{
		Use:   "docs",
		Short: "generate documentation",
	}

	docs.AddCommand(&cobra.Command{
		Use:   "cli [PATH]",
		Short: "generate CLI documentation",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := doc.GenMarkdownTree(root, args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	docs.AddCommand(cmd.APIDocs)

	root.AddCommand(docs)

	root.PersistentFlags().StringArrayP("config", "c", []string{utils.GetEnvOrDefault("PLATFORM_CONFIG", "config.yml")}, "Path to config file")
	root.PersistentFlags().StringArrayP("extra", "e", nil, "Extra arguments to apply e.g. -e ldap.domain=example.com")
	root.PersistentFlags().StringP("kubeconfig", "", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")), "Specify a kubeconfig to use, if empty a new kubeconfig is generated from master CA's at runtime")
	root.PersistentFlags().CountP("loglevel", "v", "Increase logging level")
	root.PersistentFlags().Bool("dry-run", false, "Don't apply any changes, print what would have been done")
	root.PersistentFlags().Bool("trace", false, "Print out generated specs and configs")
	root.SetUsageTemplate(root.UsageTemplate() + fmt.Sprintf("\nversion: %s\n ", version))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
