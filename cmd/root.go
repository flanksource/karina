package cmd

import (
	"fmt"
	"github.com/flanksource/commons/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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

// GetRootCmd returns the main CLI command.
func GetRootCmd(version string) *cobra.Command {
	root.AddCommand(
		Access,
		APIDocs,
		Apply,
		Backup,
		CA,
		Cleanup,
		Config,
		Conformance,
		Consul,
		DB,
		Deploy,
		Docs,
		DNS,
		Exec,
		ExecNode,
		Harbor,
		Images,
		Logs,
		MachineImages,
		NSX,
		Opa,
		Provision,
		Render,
		Rolling,
		Snapshot,
		Status,
		Test,
		Upgrade,
		Vault,
	)

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of platform-cli",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	root.PersistentFlags().StringArrayP("config", "c", []string{utils.GetEnvOrDefault("PLATFORM_CONFIG", "config.yml")}, "Path to config file")
	root.PersistentFlags().StringArrayP("extra", "e", nil, "Extra arguments to apply e.g. -e ldap.domain=example.com")
	root.PersistentFlags().CountP("loglevel", "v", "Increase logging level")
	root.PersistentFlags().Bool("dry-run", false, "Don't apply any changes, print what would have been done")
	root.PersistentFlags().Bool("trace", false, "Print out generated specs and configs")
	root.SetUsageTemplate(root.UsageTemplate() + fmt.Sprintf("\nversion: %s\n ", version))

	return root
}
