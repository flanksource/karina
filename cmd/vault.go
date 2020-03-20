package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/vault"
)

var Vault = &cobra.Command{
	Use:   "vault",
	Short: "Commands for working with vault",
}

func init() {

	init := &cobra.Command{
		Use:   "init",
		Short: "Initialize and import CA into vault",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := vault.Init(getPlatform(cmd)); err != nil {
				log.Fatalf("Failed to initialize vault %v", err)
			}
		},
	}

	Vault.AddCommand(init)
}
