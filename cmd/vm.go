package cmd

import "github.com/spf13/cobra"

var VM = &cobra.Command{
	Use:   "vm",
	Short: "Manage Virtual Machines",
}

func init() {
	tag := &cobra.Command{
		Use:   "tag",
		Short: "Tag virtual machines",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
		},
	}

	tag.Flags().String("vm", "", "The name of the VM")
	VM.AddCommand(tag)
}
