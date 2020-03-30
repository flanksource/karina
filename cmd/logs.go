package cmd

import (
	"io/ioutil"

	"github.com/moshloop/platform-cli/pkg/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Logs = &cobra.Command{
	Use:   "logs",
	Short: "Retrieve and export logs from ElasticSearch",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		kql, _ := cmd.Flags().GetString("query")
		pod, _ := cmd.Flags().GetString("pod")
		count, _ := cmd.Flags().GetInt("count")
		namespace, _ := cmd.Flags().GetString("namespace")
		cluster, _ := cmd.Flags().GetString("cluster")
		tmp, _ := ioutil.TempDir("", "logs")
		log.Infof("Exporting logs to %s", tmp)
		if err := elastic.ExportLogs(getPlatform(cmd), elastic.Query{
			Pod:       pod,
			Count:     count,
			Cluster:   cluster,
			Namespace: namespace,
			Query:     kql,
		}, tmp); err != nil {
			log.Fatalf("Failed to export logs, %s", err)
		}
	},
}

func init() {
	Logs.Flags().StringP("query", "q", "", "KQL query")
	Logs.Flags().String("cluster", "", "The kubernetes cluster to search in")
	Logs.Flags().Int("count", 1000, "Number of log entries to return")
	Logs.Flags().StringP("pod", "p", "", "Restrict to pod")
	Logs.Flags().StringP("namespace", "n", "", "Restruct to namespace")
}
