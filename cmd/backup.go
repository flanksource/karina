package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/phases/velero"
)

var Backup = &cobra.Command{
	Use:   "backup",
	Short: "Create new velero backup",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := velero.CreateBackup(getPlatform(cmd)); err != nil {
			log.Fatalf("Error creating backup %v", err)
		}
	},
}
