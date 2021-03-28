package cmd

import (
	"strings"

	pgapi "github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/karina/pkg/client/postgres"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	if superuser != "" {
		db.Superuser = superuser
	}
	return db, nil
}

func init() {
	create := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			config := pgapi.NewClusterConfig(clusterName, "test")
			config.Namespace = namespace
			config.BackupSchedule, _ = cmd.Flags().GetString("wal-schedule")
			config.EnableWalArchiving, _ = cmd.Flags().GetBool("wal-archiving")
			config.EnableWalClusterID, _ = cmd.Flags().GetBool("wal-enable-cluster-uid")
			config.UseWalgRestore, _ = cmd.Flags().GetBool("wal-use-walg-restore")

			db, err := postgresoperator.GetOrCreateDB(getPlatform(cmd), config)
			if err != nil {
				log.Fatalf("Error creating db: %v", err)
			}
			log.Infof("Created: %+v", db)
		},
	}

	create.Flags().String("wal-schedule", "*/5 * * * *", "A cron schedule to backup wal logs")
	create.Flags().Bool("wal-archiving", true, "Enable wal archiving")
	create.Flags().Bool("wal-enable-cluster-uid", false, "Enable cluster UID in wal logs s3 path")
	create.Flags().Bool("wal-use-walg-restore", true, "Enable wal-g for wal restore")
	DB.AddCommand(create)

	clone := &cobra.Command{
		Use: "clone",
		Run: func(cmd *cobra.Command, args []string) {
			config := pgapi.NewClusterConfig(clusterName, "test")
			config.Namespace = namespace
			config.BackupSchedule, _ = cmd.Flags().GetString("wal-schedule")
			config.EnableWalArchiving, _ = cmd.Flags().GetBool("wal-archiving")
			config.EnableWalClusterID, _ = cmd.Flags().GetBool("wal-enable-cluster-uid")
			config.UseWalgRestore, _ = cmd.Flags().GetBool("wal-use-walg-restore")
			cloneClusterName, _ := cmd.Flags().GetString("clone-cluster-name")
			cloneTimestamp, _ := cmd.Flags().GetString("clone-timestamp")
			config.Clone = &pgapi.CloneConfig{
				ClusterName: cloneClusterName,
				Timestamp:   cloneTimestamp,
			}
			if config.EnableWalClusterID {
				config.Clone.ClusterID, _ = cmd.Flags().GetString("clone-cluster-uid")
			}

			db, err := postgresoperator.GetOrCreateDB(getPlatform(cmd), config)
			if err != nil {
				log.Fatalf("Error creating db: %v", err)
			}
			log.Infof("Created: %+v", db)
		},
	}

	clone.Flags().String("wal-schedule", "*/5 * * * *", "A cron schedule to backup wal logs")
	clone.Flags().Bool("wal-archiving", true, "Enable wal archiving")
	clone.Flags().Bool("wal-enable-cluster-uid", false, "Enable cluster UID in wal logs s3 path")
	clone.Flags().Bool("wal-use-walg-restore", true, "Enable wal-g for wal restore")
	clone.Flags().String("clone-cluster-name", "", "Name of the cluster to clone")
	clone.Flags().String("clone-cluster-uid", "", "UID of the cluster to clone")
	clone.Flags().String("clone-timestamp", "", "Timestamp of the wal to clone")
	DB.AddCommand(clone)

	DB.AddCommand(&cobra.Command{
		Use:   "restore [backup bucket] <backup path>",
		Short: "Restore a database from backups",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			db, err := getDB(cmd)
			if err != nil {
				log.Fatalf("error finding %s: %v", clusterName, err)
			}
			log.Infof("Restoring %s from %s", db, strings.Join(args, " "))
			if err := db.Restore(args...); err != nil {
				log.Fatalf("Error Restore up db %s\n", err)
			}
		},
	})

	backup := &cobra.Command{
		Use:   "backup",
		Short: "Create a new database backup",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			listBackups, _ := cmd.Flags().GetBool("list")
			schedule, _ := cmd.Flags().GetString("schedule")
			db, err := getDB(cmd)
			if err != nil {
				log.Fatalf("error finding %s: %v", clusterName, err)
			}

			if listBackups {
				s3Bucket, _ := cmd.Flags().GetString("bucket")
				log.Infof("Querying for list of snapshot for %s", db)
				if err := db.ListBackups(s3Bucket); err != nil {
					log.Fatalf("Failed to list backups: %v", err)
				}
			} else if schedule != "" {
				log.Infof("Creating backup schedule: %s: %s", schedule, db)
				if err := db.ScheduleBackup(schedule); err != nil {
					log.Fatalf("Failed to create backup schedule: %v", err)
				}
			} else {
				log.Infof("Backing up %s", db)

				if err := db.Backup(); err != nil {
					log.Fatalf("Error backing up db %s\n", err)
				}
			}
		},
	}

	backup.Flags().Bool("list", false, "List all backup revisions")
	backup.Flags().String("bucket", "", "List all backup revisions in a specific bucket")
	backup.Flags().String("schedule", "", "A cron schedule to backup on a reoccuring basis")
	DB.AddCommand(backup)

	DB.PersistentFlags().StringVar(&clusterName, "name", "", "Name of the postgres cluster / service")
	DB.PersistentFlags().StringVar(&namespace, "namespace", "postgres-operator", "")
	DB.PersistentFlags().StringVar(&secret, "secret", "", "Name of the secret that contains the postgres user credentials")
	DB.PersistentFlags().StringVar(&superuser, "superuser", "", "Superuser user")
}
