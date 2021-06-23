package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/flanksource/karina/cmd"
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
		cmd.BurninController,
		cmd.CA,
		cmd.Terminate,
		cmd.CleanupJobs,
		cmd.Config,
		cmd.Conformance,
		cmd.Consul,
		cmd.Dashboard,
		cmd.DB,
		cmd.Deploy,
		cmd.DNS,
		cmd.Exec,
		cmd.ExecNode,
		cmd.Etcd,
		cmd.Harbor,
		cmd.Images,
		cmd.Logs,
		cmd.MachineImages,
		cmd.Namespace,
		cmd.Node,
		cmd.NSX,
		cmd.Operator,
		cmd.Orphan,
		cmd.Provision,
		cmd.Render,
		cmd.Report,
		cmd.Rolling,
		cmd.Snapshot,
		cmd.Status,
		cmd.Test,
		cmd.TerminateNodes,
		cmd.TerminateOrphans,
		cmd.Undelete,
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

	config := "karina.yml"
	if env := os.Getenv("PLATFORM_CONFIG"); env != "" {
		config = env
	}

	root.PersistentFlags().StringArrayP("config", "c", []string{config}, "Path to config file")
	root.PersistentFlags().StringArrayP("extra", "e", nil, "Extra arguments to apply e.g. -e ldap.domain=example.com")
	root.PersistentFlags().StringP("kubeconfig", "", "", "Specify a kubeconfig to use, if empty a new kubeconfig is generated from master CA's at runtime")
	root.PersistentFlags().CountP("loglevel", "v", "Increase logging level")
	root.PersistentFlags().Bool("prune", true, "Delete previously enabled resources")
	root.PersistentFlags().Bool("skip-decrypt", false, "Skip any phases requiring decryption")
	root.PersistentFlags().Bool("dry-run", false, "Don't apply any changes, print what would have been done")
	root.PersistentFlags().Bool("trace", false, "Print out generated specs and configs")
	root.PersistentFlags().Bool("in-cluster", false, "Use in cluster kubernetes config")
	root.SetUsageTemplate(root.UsageTemplate() + fmt.Sprintf("\nversion: %s\n ", version))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
