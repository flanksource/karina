package cmd

import (
	"github.com/moshloop/platform-cli/pkg/reports"
	"github.com/spf13/cobra"
)

var Report = &cobra.Command{
	Use: "report",
}

var reportOpts = reports.ReportOptions{}

func init() {
	Report.PersistentFlags().StringVar(&reportOpts.Path, "input", "", "Path to input directory of specs")
	Report.PersistentFlags().StringArrayVar(&reportOpts.Annotations, "col", nil, "Annotations to include in the report")
	Report.PersistentFlags().StringVar(&reportOpts.Format, "format", "table", "Format of the report, can be one of table,csv")
	Report.AddCommand(&cobra.Command{
		Use: "quotas",
		RunE: func(cmd *cobra.Command, args []string) error {
			return reports.Quotas(reportOpts)
		},
	})
}
