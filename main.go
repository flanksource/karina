package main

import (
	"fmt"
	"github.com/spf13/cobra/doc"
	"os"

	"github.com/moshloop/platform-cli/cmd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			case level > 2:
				log.SetLevel(log.TraceLevel)
			case level > 1:
				log.SetLevel(log.DebugLevel)
			case level > 0:
				log.SetLevel(log.InfoLevel)
			default:
				log.SetLevel(log.WarnLevel)
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
		cmd.Access)

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

	root.PersistentFlags().StringP("config", "c", "config.yml", "Path to config file")
	root.PersistentFlags().CountP("loglevel", "v", "Increase logging level")
	root.SetUsageTemplate(root.UsageTemplate() + fmt.Sprintf("\nversion: %s\n ", version))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
