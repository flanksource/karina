package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/harbor"
)

// Harbor is the parent command for interactor with the harbor docker registry
var Harbor = &cobra.Command{
	Use:   "harbor",
	Short: "Commmands for deploying and interacting with harbor",
}

func init() {

	Harbor.AddCommand(&cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy the harbor registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

	Harbor.AddCommand(&cobra.Command{
		Use:   "backup",
		Short: "Backup the harbor database",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.Backup(getPlatform(cmd)); err != nil {
				log.Fatalf("Error backing up harbor %s\n", err)
			}
		},
	})

	Harbor.AddCommand(&cobra.Command{
		Use:   "restore",
		Short: "Restore the harbor database from backups",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.Restore(getPlatform(cmd), ""); err != nil {
				log.Fatalf("Error restoring harbor %s\n", err)
			}
		},
	})
	Harbor.AddCommand(&cobra.Command{
		Use:   "replicate-all",
		Short: "Trigger a manual replication for all enabled jobs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.ReplicateAll(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

}
