package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Delete = &cobra.Command{
	Use:   "delete",
	Short: "delete all pods",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("hello world")
  },
}
