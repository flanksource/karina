package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/commons/exec"
)

var Images = &cobra.Command{
	Use:   "images",
	Short: "Commands for working with docker images",
}

func init() {
	Images.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all docker images used by the platform",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			images := []string{}
			for _, prefix := range []string{"image:", "image=="} {
				stdout, ok := exec.SafeExec("cat build/*.yaml | grep -i %s | cut -d: -f2 -f3 | sort | uniq", prefix)
				if !ok {
					log.Fatalf("Failed to list images %s", stdout)
				}
				for _, line := range strings.Split(stdout, "\n") {
					if line == "" {
						continue
					}
					images = append(images, strings.Trim(line, " "))
				}
			}
			fmt.Println(strings.Join(images, "\n"))

		},
	})
	Images.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize all platform docker images to a local registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

		},
	})
}
