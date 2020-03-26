package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "Commands for working with config files",
}

var validateConfig = &cobra.Command{
	Use:   "validate",
	Short: "Validate config",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p := getPlatform(cmd)
		config := p.String()
		fmt.Printf("Generated config is:\n%s\n", config)
	},
}

func init() {
	Config.AddCommand(validateConfig)
}
