package cmd

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/spf13/cobra"
)

var Orphan = &cobra.Command{
	Use:   "orphan",
	Short: "Remove owner references from an object",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		platform := getPlatform(cmd)

		kind := args[0]
		name := args[1]
		namespace, _ := cmd.Flags().GetString("namespace")

		object, found := constants.RuntimeObjects[kind]
		if !found {
			if err := platform.OrphanCRD(kind, name, namespace); err != nil {
				platform.Fatalf("failed to orphan %s %s in namespace %s: %v", kind, name, namespace, err)
			}
		} else {
			if err := platform.Orphan(kind, name, namespace, object); err != nil {
				platform.Fatalf("failed to orphan %s %s in namespace %s: %v", kind, name, namespace, err)
			}
		}
	},
}

func init() {
	Orphan.Flags().StringP("namespace", "n", "", "Namespace where object is present")
}
