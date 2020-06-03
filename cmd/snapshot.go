package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/phases/snapshot"
)

var opts = snapshot.Options{}
var Snapshot = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a snapshot of the running system",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		opts.Namespaces = args
		if err := snapshot.Take(getPlatform(cmd), opts); err != nil {
			log.Fatalf("Failed to get cluster snapshot, %s", err)
		}
	},
}

func init() {
	Snapshot.Flags().StringVarP(&opts.Destination, "output-dir", "o", "snapshot", "Output directory for snapshot")
	since, _ := time.ParseDuration("4h")
	Snapshot.Flags().DurationVar(&opts.LogsSince, "since", since, "Return logs newer than a relative duration like 5s, 2m, or 3h")
	Snapshot.Flags().BoolVarP(&opts.IncludeSpecs, "include-specs", "", false, "Export yaml specs")
	Snapshot.Flags().BoolVarP(&opts.IncludeEvents, "include-events", "", true, "Export events")
	Snapshot.Flags().BoolVarP(&opts.IncludeLogs, "include-logs", "", true, "Export logs for pods")
	Snapshot.Flags().IntVar(&opts.Concurrency, "concurrency", 1, "Run the export concurrently")
}
