package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/phases"
	"github.com/spf13/cobra"
)

var Access = &cobra.Command{
	Use:   "access",
	Short: "Cleanup a cluster and terminate all VM's and resources associated with it",
}

func init() {
	Access.AddCommand(&cobra.Command{
		Use:   "kubeconfig",
		Short: "Generate a new kubeconfig file for accessing the cluster",
		Run: func(cmd *cobra.Command, args []string) {

			platform := getPlatform(cmd)
			ips := platform.GetMasterIPs()
			data, err := phases.CreateKubeConfig(platform, ips[0])
			if err != nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	})

	Access.AddCommand(&cobra.Command{
		Use:   "sso",
		Short: "Generate a new kubeconfig file for accessing the cluster using sso",
		Run: func(cmd *cobra.Command, args []string) {

			platform := getPlatform(cmd)
			ips := platform.GetMasterIPs()
			data, err := phases.CreateOIDCKubeConfig(platform, ips[0])
			if err != nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	})
}
