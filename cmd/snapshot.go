package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/snapshot"
)

var outputDir string
var since time.Duration
var Snapshot = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a snapshot of the running system",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := snapshot.Take(getPlatform(cmd), outputDir, since); err != nil {
			log.Fatalf("Failed to get cluster snapshot, %s", err)
		}
	},
}

func init() {
	Snapshot.Flags().StringVarP(&outputDir, "output-dir", "o", "snapshot", "Output directory for snapshot")
	since, _ := time.ParseDuration("4h")
	Snapshot.Flags().DurationVar(&since, "since", since, "Return logs newer than a relative duration like 5s, 2m, or 3h")
}
