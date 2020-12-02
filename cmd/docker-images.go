package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/commons/exec"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/flanksource/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var Images = &cobra.Command{
	Use:   "images",
	Short: "Commands for working with docker images",
}

func getImages(p *platform.Platform) ([]string, error) {
	images := []string{}
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
			return nil, errors.Wrapf(err, "error deploying %s", name)
		}
	}
	return images, nil
}

func uniqueStrings(list []string) {
	if len(list) < 2 {
		return
	}

	sort.Strings(list)
	newList := []string{list[0]}
	length := len(list)

	for i := 1; i < length; i++ {
		if list[i] != list[i-1] {
			newList = append(newList, list[i])
		}
	}

	list = newList
}

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all docker images used by the platform",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			outputFormat, _ := cmd.Flags().GetString("output")
			p := getPlatform(cmd)
			images, err := getImages(p)
			if err != nil {
				p.Fatalf("Failed to dry-run deploy : %v", err)
			}
			if p.DockerRegistry != "" {
				l := len(images)
				for i := 0; i < l; i++ {
					if !strings.HasPrefix(images[i], p.DockerRegistry) {
						images[i] = p.DockerRegistry + "/" + images[i]
					}
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

	var imagesToSync []string
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize all platform docker images to a local registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			if len(imagesToSync) == 0 {
				images, err := getImages(p)
				imagesToSync = images
				if err != nil {
					p.Fatalf("Failed to dry-run deploy : %v", err)
				}
			}

			uniqueStrings(imagesToSync)

			for _, image := range imagesToSync {
				if strings.HasPrefix(image, p.DockerRegistry) {
					p.Infof("Image %s is used from target registry", image)
					continue
				}
				if err := harbor.CheckManifest(p, image); err != nil {
					p.Infof("Syncing %s", image)
					_ = exec.Execf("docker pull %s", image)
					_ = exec.Execf("docker tag %s %s/%s", image, p.DockerRegistry, image)
					_ = exec.Execf("docker push %s/%s", p.DockerRegistry, image)
				} else {
					p.Infof("Image %s already exists in target repository", image)
				}
			}
		},
	}
	syncCmd.Flags().StringArrayVarP(&imagesToSync, "image", "i", []string{}, "A list of images to sync")
	Images.AddCommand(listCmd, syncCmd)
}
