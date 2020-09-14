package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/flanksource/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var Images = &cobra.Command{
	Use:   "images",
	Short: "Commands for working with docker images",
}

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all docker images used by the platform",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			images := []string{}

			outputFormat, _ := cmd.Flags().GetString("output")

			p := getPlatform(cmd)
			// in order to list all images we perform an dry-run deployment
			// with an ApplyHook
			p.DryRun = true
			p.ApplyDryRun = true
			p.TerminationProtection = true
			p.ApplyHook = func(ns string, obj unstructured.Unstructured) {
				containers := []interface{}{}
				if image, found := obj.GetAnnotations()["image"]; found {
					images = append(images, image)
				}
				list, found, _ := unstructured.NestedSlice(obj.UnstructuredContent(), "spec", "template", "spec", "containers")
				if found {
					containers = append(containers, list...)
				}
				list, found, _ = unstructured.NestedSlice(obj.UnstructuredContent(), "spec", "template", "spec", "initContainers")
				if found {
					containers = append(containers, list...)
				}
				for _, container := range containers {
					image, found := container.(map[string]interface{})["image"].(string)
					if found {
						images = append(images, image)
					}
				}
			}

			for name, fn := range Phases {
				if err := fn(p); err != nil {
					log.Errorf("Failed to dry-run deploy %s: %v", name, err)
				}
			}

			if outputFormat == "text" {
				fmt.Println(strings.Join(images, "\n"))
			} else if outputFormat == "yaml" {
				yml, _ := yaml.Marshal(images)
				fmt.Println(string(yml))
			}
		},
	}

	listCmd.PersistentFlags().StringP("output", "o", "text", "Output format (string, yaml)")
	Images.AddCommand(listCmd)

	Images.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize all platform docker images to a local registry",
		Args:  cobra.MinimumNArgs(0),
		Run:   func(cmd *cobra.Command, args []string) {},
	})
}
