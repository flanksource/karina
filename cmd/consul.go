package cmd

import (
	"github.com/moshloop/platform-cli/pkg/phases/consul"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Consul = &cobra.Command{
	Use: "consul",
}

func init() {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a new consul backup",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			namespace, _ := cmd.Flags().GetString("namespace")
			schedule, _ := cmd.Flags().GetString("schedule")

			cs := consul.NewBackupRestore(getPlatform(cmd), name, namespace)

			if schedule != "" {
				log.Infof("Creating consul backup schedule: %s: %s", schedule, cs.Name)
				if err := cs.ScheduleBackup(schedule); err != nil {
					log.Fatalf("Failed to create backup schedule: %v", err)
				}
			} else {
				log.Infof("Backing up consul %s", cs.Name)

				if err := cs.Backup(); err != nil {
					log.Fatalf("Error backing up consul %s: %v\n", cs.Name, err)
				}
			}
		},
	}
	backupCmd.Flags().String("name", "", "Name of the consul deployment")
	backupCmd.Flags().String("namespace", "", "Namespace where consul runs")
	backupCmd.Flags().String("schedule", "", "A cron schedule to backup on a recurring basis")

	restoreCmd := &cobra.Command{
		Use:   "restore [backup path]",
		Short: "Restore consul from backups",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			namespace, _ := cmd.Flags().GetString("namespace")

			cs := consul.NewBackupRestore(getPlatform(cmd), name, namespace)

			log.Infof("Restoring consul %s from %s", cs.Name, args[0])
			if err := cs.Restore(args[0]); err != nil {
				log.Fatalf("Error Restore up consul %s: %v\n", cs.Name, err)
			}
		},
	}
	restoreCmd.Flags().String("name", "", "Name of the consul deployment")
	restoreCmd.Flags().String("namespace", "", "Namespace where consul runs")

	Consul.AddCommand(backupCmd)
	Consul.AddCommand(restoreCmd)
}
