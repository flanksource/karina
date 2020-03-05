package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/db"
)

var clusterName string
var DB = &cobra.Command{
	Use: "db",
}

func init() {
	DB.AddCommand(&cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			db, err := platform.GetOrCreateDB(clusterName, "test")
			if err != nil {
				log.Fatalf("Error creating db: %v", err)
			}
			log.Infof("Created: %+v", db)
		},
	})
	DB.AddCommand(&cobra.Command{
		Use:   "restore",
		Short: "Restore a database from backups",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := db.Backup(getPlatform(cmd), clusterName); err != nil {
				log.Fatalf("Error backuping up db %s\n", err)
			}
		},
	})

	DB.AddCommand(&cobra.Command{
		Use:   "backup",
		Short: "Create a new database backup",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := db.Restore(getPlatform(cmd), clusterName, ""); err != nil {
				log.Fatalf("Error restoring db %s\n", err)
			}
		},
	})

	DB.PersistentFlags().StringVar(&clusterName, "name", "", "")

}
