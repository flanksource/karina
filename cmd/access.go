package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/k8s"
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

			endpoint, err := platform.GetAPIEndpoint()
			if err != nil {
				log.Fatalf("Unable to get API endpoint: %v", err)
			}
			group, _ := cmd.Flags().GetString("group")
			name, _ := cmd.Flags().GetString("name")
			expiry, _ := cmd.Flags().GetDuration("expiry")
			ca, err := platform.GetCA()
			if err != nil {
				log.Fatalf("Error getting CA: %v", err)
			}

			data, err := k8s.CreateKubeConfig(platform.Name, ca, endpoint, group, name, expiry)
			if err != nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	}

	admin.Flags().String("name", "kubernetes-admin", "The name to use in the certificate")
	admin.Flags().String("group", "system:masters", "The OU (group name) to use in the certificate")
	admin.Flags().Duration("expiry", 24*7*time.Hour, "Validity in days of the certificate")
	Access.AddCommand(admin)

	Access.AddCommand(&cobra.Command{
		Use:   "sso",
		Short: "Generate a new kubeconfig file for accessing the cluster using sso",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			ca, err := platform.GetCA()
			if err != nil {
				log.Fatalf("Error getting CA: %v", err)
			}
			data, = k8s.CreateOIDCKubeConfig(platform.Name, ca, fmt.Sprintf("k8s-api.%s", platform.Domain), fmt.Sprintf("dex.%s", platform.Domain), "", "", "")
			if err != err :nil {
				log.Fatalf("Failed to create kubeconfig %s", err)
			}
			fmt.Println(string(data))
		},
	})
}
