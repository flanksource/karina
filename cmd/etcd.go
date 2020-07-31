package cmd

import (
	"github.com/flanksource/karina/pkg/provision"
	"github.com/spf13/cobra"
)

var Etcd = &cobra.Command{
	Use: "etcd",
}

func init() {
	Etcd.AddCommand(&cobra.Command{
		Use: "status",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			client, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("could not get client: %v", err)
			}
			if err := provision.GetEtcdClient(platform, client).PrintStatus(); err != nil {
				platform.Fatalf("Failed to get etcd status: %v", err)
			}
		},
	})

	Etcd.AddCommand(&cobra.Command{
		Use:  "remove-member [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			client, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("could not get client: %v", err)
			}
			provision.GetEtcdClient(platform, client).RemoveMember(args[0])
		},
	})

	Etcd.AddCommand(&cobra.Command{
		Use:  "move-leader [name]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			client, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("could not get client: %v", err)
			}
			provision.GetEtcdClient(platform, client).MoveLeader(args[0])
		},
	})

}
