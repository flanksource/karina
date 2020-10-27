package cmd

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/spf13/cobra"
)

var Undelete = &cobra.Command{
	Use:   "undelete",
	Short: "Undelete kubernetes objects",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		platform := getPlatform(cmd)

		kind := args[0]
		name := args[1]
		namespace, _ := cmd.Flags().GetString("namespace")

		object, found := constants.RuntimeObjects[kind]
		if !found {
			platform.Fatalf("kind %s not supported", kind)
		}

		if err := platform.Undelete(kind, name, namespace, object); err != nil {
			platform.Fatalf("failed to undelete %s %s in namespace %s: %v", kind, name, namespace, err)
		}
	},
}

func init() {
	Undelete.Flags().StringP("namespace", "n", "", "Namespace where object is present")
}
