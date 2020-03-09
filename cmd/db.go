package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/client/postgres"
)

var DB = &cobra.Command{
	Use: "db",
}

var clusterName, namespace, secret, superuser string

func getDB(cmd *cobra.Command) (*postgres.PostgresDB, error) {
	platform := getPlatform(cmd)
	s3, err := platform.GetS3Client()
	if err != nil {
		return nil, err
	}
	var db *postgres.PostgresDB
	if namespace == "postgres-operator" {
		db, err = postgres.GetPostgresDB(&platform.Client, s3, clusterName)
	} else {
		db, err = postgres.GetGenericPostgresDB(&platform.Client, s3, namespace, clusterName, secret, "12")
	}
	if err != nil {
		return nil, err
	}
	if secret != "" {
		db.Secret = secret
	}
	db.Superuser = superuser
	return db, nil
}

func init() {

	DB.AddCommand(&cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {

			db, err := getPlatform(cmd).GetOrCreateDB(clusterName, "test")
			if err != nil {
				log.Fatalf("Error creating db: %v", err)
			}
			log.Infof("Created: %+v", db)
		},
	})
	DB.AddCommand(&cobra.Command{
		Use:   "restore [backup path]",
		Short: "Restore a database from backups",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			db, err := getDB(cmd)
			if err != nil {
				log.Fatalf("error finding %s: %v", clusterName, err)
			}
			log.Infof("Restoring %s from %s", db, args[0])
			if err := db.Restore(args[0]); err != nil {
				log.Fatalf("Error Restore up db %s\n", err)
			}
		},
	})

	DB.AddCommand(&cobra.Command{
		Use:   "backup",
		Short: "Create a new database backup",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			db, err := getDB(cmd)
			if err != nil {
				log.Fatalf("error finding %s: %v", clusterName, err)
			}
			log.Infof("Backing up %s", db)

			if err := db.Backup(); err != nil {
				log.Fatalf("Error backing up db %s\n", err)
			}
		},
	})

	DB.PersistentFlags().StringVar(&clusterName, "name", "", "Name of the postgres cluster / service")
	DB.PersistentFlags().StringVar(&namespace, "namespace", "postgres-operator", "")
	DB.PersistentFlags().StringVar(&secret, "secret", "", "Name of the secret that contains the postgres user credentials")
	DB.PersistentFlags().StringVar(&superuser, "superuser", "postgres", "Superuser user")

}
