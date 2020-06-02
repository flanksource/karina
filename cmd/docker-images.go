package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
			p := getPlatform(cmd)
			// in order to list all images we perform an dry-run deployment
			// with an ApplyHook
			p.DryRun = true
			p.ApplyDryRun = true
			p.TerminationProtection = true
			p.ApplyHook = func(ns string, obj unstructured.Unstructured) {
				containers := []interface{}{}
				if image, found := obj.GetAnnotations()["image"]; found {
					fmt.Println(image)
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
					image, found := container.(map[string]interface{})["image"]
					if found {
						fmt.Printf("%s/%s\n", obj.GetName(), image)
					}
				}
			}
			for name, fn := range Phases {
				if err := fn(p); err != nil {
					log.Errorf("Failed to dry-run deploy %s: %v", name, err)
				}
			}

			fmt.Println(strings.Join(images, "\n"))
		},
	})

	Images.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Synchronize all platform docker images to a local registry",
		Args:  cobra.MinimumNArgs(0),
		Run:   func(cmd *cobra.Command, args []string) {},
	})
}
