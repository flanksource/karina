package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/k8s"
)

var Access = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Generate kubeconfig files",
}

func init() {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Generate a new kubeconfig file for accessing the cluster using an X509 Certificate",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			ips := platform.GetMasterIPs()

			var endpoint string
			if platform.DNS != nil && !platform.DNS.Disabled {
				endpoint = fmt.Sprintf("k8s-api.%s", platform.Domain)
			} else {
				// No DNS available using the first masters IP as an endpoint
				endpoint = ips[0]
			}
			group, _ := cmd.Flags().GetString("group")
			name, _ := cmd.Flags().GetString("name")
			data, err := k8s.CreateKubeConfig(platform.Name, platform.GetCA(), endpoint, group, name)
			if err != nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	}

	admin.Flags().String("name", "kubernetes-admin", "The name to use in the certificate")
	admin.Flags().String("group", "system:masters", "The OU (group name) to use in the certificate")

	Access.AddCommand(admin)

	Access.AddCommand(&cobra.Command{
		Use:   "sso",
		Short: "Generate a new kubeconfig file for accessing the cluster using sso",
		Run: func(cmd *cobra.Command, args []string) {

			platform := getPlatform(cmd)
			ips := platform.GetMasterIPs()
			data, err := k8s.CreateOIDCKubeConfig(platform.Name, platform.GetCA(), ips[0], fmt.Sprintf("https://dex.%s", platform.Domain))
			if err != nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	})
}
