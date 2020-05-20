package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/provision"
)

var Cleanup = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"terminate"},
	Short:   "Terminate a cluster and destroy all VM's and resources associated with it",
	Args:    cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Cleanup(getPlatform(cmd)); err != nil {
			log.Fatalf("Failed to cleanup cluster, %s", err)
		}
	},
}

var TerminateNodes = &cobra.Command{
	Use:   "terminate-node [nodes]",
	Short: "Cordon and terminate the specified nodes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.TerminateNodes(getPlatform(cmd), args); err != nil {
			log.Fatalf("Failed terminate nodes %s", err)
		}
	},
}

var TerminateOrphans = &cobra.Command{
	Use:   "terminate-orphans",
	Short: "Terminate all orphaned VM's that have not successfully joined the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.TerminateOrphans(getPlatform(cmd)); err != nil {
			log.Fatalf("Failed terminate nodes %s", err)
		}
	},
}
