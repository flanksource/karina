package cmd

import (
	"fmt"

	"github.com/hako/durafmt"

	"github.com/spf13/cobra"

	"time"
)

var Namespace = &cobra.Command{
	Use:   "namespace",
	Short: "Commands for manipulating namespaces",
}

var timeout time.Duration

var namespaceForceDelete = &cobra.Command{
	Use:   "force-delete",
	Short: "Clears the namespace finalizers allowing it to be deleted",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := getPlatform(cmd)

		ns := args[0]
		fmt.Printf("Clearing finalizers on namespace %v\n", ns)

		duration := durafmt.Parse(timeout).String()
		fmt.Printf("Using timeout of %v\n", duration)

		err := p.ForceDeleteNamespace(ns, timeout)
		if err != nil {
			return fmt.Errorf("clearing the finalizers failed with: %v", err)
		}

		return nil
	},
}

func init() {
	namespaceForceDelete.Flags().DurationVar(&timeout, "timeout", time.Minute*2, "specify the timeout")
	Namespace.AddCommand(namespaceForceDelete)
}
