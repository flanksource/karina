package cmd

import (
	"path"

	"github.com/flanksource/karina/pkg/phases/monitoring"
	"github.com/spf13/cobra"
)

var Dashboard = &cobra.Command{
	Use: "dashboard",
}

func init() {
	deploy := &cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			file, _ := cmd.Flags().GetString("file")
			name, _ := cmd.Flags().GetString("name")
			folder, _ := cmd.Flags().GetString("folder")
			if name == "" {
				name = path.Base(file)
			}
			if err := monitoring.DeployDashboard(p, folder, name, file); err != nil {
				p.Fatalf("failed to deploy dashboard: %v", err)
			}
			p.Infof("Dashboard %s deployed !", name)
		},
	}

	deploy.Flags().StringP("file", "f", "", "Grafana dashboard which needs to be deployed")
	deploy.Flags().StringP("name", "n", "", "Grafana dashboard name")
	deploy.Flags().StringP("folder", "", "Customer", "Grafana dashboard folder name")
	Dashboard.AddCommand(deploy)
}
