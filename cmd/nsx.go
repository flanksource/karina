package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/nsx"
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
			if err := nsx.SetLogLevel(getPlatform(cmd), level); err != nil {
				log.Errorf("Failed to set nsx logging level %v", err)
			}
		},
	}

	logLevel.Flags().String("log-level", "WARNING", fmt.Sprintf("Update the log level to one of %v", nsx.LogLevels))
	NSX.AddCommand(logLevel)
}
