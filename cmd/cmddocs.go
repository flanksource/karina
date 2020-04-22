package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var Docs = &cobra.Command{
	Use:   "docs",
	Short: "generate documentation",
}

func init() {
	Docs.AddCommand(&cobra.Command{
		Use:   "cli [PATH]",
		Short: "generate CLI documentation",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := doc.GenMarkdownTree(&rootCmd, args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	Docs.AddCommand(APIDocs)
}
