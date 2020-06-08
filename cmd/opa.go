package cmd

import (
	"github.com/flanksource/karina/pkg/phases/opa"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Opa = &cobra.Command{
	Use: "opa",
}

func init() {
	Opa.AddCommand(&cobra.Command{
		Use:   "deploy-bundle",
		Short: "deploy opa bundle",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opa.DeployBundle(getPlatform(cmd), args[0]); err != nil {
				log.Fatalf("Error deploying  opa bundles: %s", err)
			}
		},
	})
}
