package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Operator = &cobra.Command{
	Use:   "operator",
	Short: "Run karina operator",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		op, err := operator.NewOperator()
		if err != nil {
			log.Fatalf("failed to create operator: %v", err)
		}

		if err := op.Run(); err != nil {
			log.Fatalf("failed to start operator: %v", err)
		}
	},
}

func init() {
}
