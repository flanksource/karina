package cmd

import (
	"github.com/flanksource/karina/pkg/provision"
	"github.com/spf13/cobra"
)

var Etcd = &cobra.Command{
	Use: "etcd",
}

var etcdHost string

func init() {
	Etcd.PersistentFlags().StringVar(&etcdHost, "etcd-host", "", "Etcd hostname to connect through")
	Etcd.AddCommand(&cobra.Command{
		Use: "status",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			client, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("could not get client: %v", err)
			}
			if err := provision.GetEtcdClient(platform, client, etcdHost).PrintStatus(); err != nil {
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
			if err := provision.GetEtcdClient(platform, client, etcdHost).RemoveMember(args[0]); err != nil {
				platform.Fatalf(err.Error())
			}
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
			if err := provision.GetEtcdClient(platform, client, etcdHost).MoveLeader(args[0]); err != nil {
				platform.Fatalf(err.Error())
			}
		},
	})
}
