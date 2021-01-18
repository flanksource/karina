package cmd

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Apply = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration to a resource by filename",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		ns, _ := cmd.Flags().GetString("namespace")
		p := getPlatform(cmd)
		for _, spec := range args {
			data, err := ioutil.ReadFile(spec)
			if err != nil {
				log.Fatalf("Could not read %s: %v", spec, err)
			}
			template, err := p.TemplateText(string(data))
			if err != nil {
				log.Fatalf("failed to template %s: %v", spec, err)
			}
			if err := p.ApplyText(ns, template); err != nil {
				log.Fatalf("failed to apply %s: %v", spec, err)
			}
		}
	},
}

func init() {
	Apply.Flags().StringP("namespace", "n", "", "The namespace to apply")
}
