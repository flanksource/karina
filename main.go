package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/moshloop/commons/utils"
	"github.com/moshloop/platform-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var root = &cobra.Command{
		Use: "platform-cli",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, _ := cmd.Flags().GetCount("loglevel")
			switch {
			case level > 1:
				log.SetLevel(log.TraceLevel)
			case level > 0:
				log.SetLevel(log.DebugLevel)
			default:
				log.SetLevel(log.InfoLevel)
			}
		},
	}

	root.AddCommand(
		cmd.Dependencies,
		cmd.Images,
		cmd.MachineImages,
		cmd.Upgrade,
		cmd.Test,
		cmd.Build,
		cmd.Provision,
		cmd.Cleanup,
		cmd.Status,
		cmd.Access,
		cmd.Deploy,
		cmd.Harbor,
		cmd.DNS,
		cmd.Render,
		cmd.Snapshot,
		cmd.Conformance)

	if len(commit) > 8 {
		version = fmt.Sprintf("%v, commit %v, built at %v", version, commit[0:8], date)
	}
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of platform-cli",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "docs",
		Short: "generate documentation",
		Run: func(cmd *cobra.Command, args []string) {
			err := doc.GenMarkdownTree(root, "docs")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Documentation generated at: docs")
		},
	})

	root.PersistentFlags().StringArrayP("config", "c", []string{utils.GetEnvOrDefault("PLATFORM_CONFIG", "config.yml")}, "Path to config file")
	root.PersistentFlags().StringArrayP("extra", "e", nil, "Extra arguments to apply e.g. -e ldap.domain=example.com")
	root.PersistentFlags().CountP("loglevel", "v", "Increase logging level")
	root.PersistentFlags().Bool("dry-run", false, "Don't apply any changes, print what would have been done")
	root.SetUsageTemplate(root.UsageTemplate() + fmt.Sprintf("\nversion: %s\n ", version))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
