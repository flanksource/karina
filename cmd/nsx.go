package cmd

// TODO: Remove this module

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var NSX = &cobra.Command{
	Use:   "nsx",
	Short: "Commands for interacting with NSX clusters",
}

func init() {
	logLevel := &cobra.Command{
		Use:   "set-log-level",
		Short: "Update the logging level",
		Run: func(cmd *cobra.Command, args []string) {
			level, _ := cmd.Flags().GetString("log-level")
			log.Infof("Setting log level to %s", level)
		},
	}

	NSX.AddCommand(logLevel)
}
