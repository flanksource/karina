package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/phases/harbor"
)

// Harbor is the parent command for interactor with the harbor docker registry
var Harbor = &cobra.Command{
	Use:   "harbor",
	Short: "Commmands for deploying and interacting with harbor",
}

func init() {
	Harbor.AddCommand(&cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy the harbor registry",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.Deploy(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

	Harbor.AddCommand(&cobra.Command{
		Use:   "update-settings",
		Short: "Update harbor settings",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.UpdateSettings(getPlatform(cmd)); err != nil {
				log.Fatalf("Error backing up harbor %s\n", err)
			}
		},
	})

	Harbor.AddCommand(&cobra.Command{
		Use:   "replicate-all",
		Short: "Trigger a manual replication for all enabled jobs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := harbor.ReplicateAll(getPlatform(cmd)); err != nil {
				log.Fatalf("Error building harbor %s\n", err)
			}
		},
	})

	list := &cobra.Command{
		Use:   "list",
		Short: "Commands to list objects",
	}
	list.AddCommand()
	Harbor.AddCommand(list)

	listProjects := &cobra.Command{
		Use:   "projects",
		Short: "List projects in harbor",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			projects, err := harbor.ListProjects(platform)
			if err != nil {
				log.Fatalf("Error listing projects: %v", err)
			}

			for _, project := range projects {
				fmt.Println(project.Name)
			}
		},
	}

	listImages := &cobra.Command{
		Use:   "images",
		Short: "List images in harbor",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			listTags, _ := cmd.Flags().GetBool("tags")
			concurrency, _ := cmd.Flags().GetInt("concurrency")

			if listTags {
				tags, err := harbor.ListImagesWithTags(platform, concurrency)
				if err != nil {
					log.Fatalf("Error listing images: %v", err)
				}

				for _, tag := range tags {
					fmt.Printf("%s/%s:%s\n", tag.ProjectName, tag.RepositoryName, tag.Digest)
				}
			} else {
				images, err := harbor.ListImages(platform, concurrency)
				if err != nil {
					log.Fatalf("Error listing images: %v", err)
				}

				for _, image := range images {
					fmt.Printf("%s/%s\n", image.ProjectName, image.Name)
				}
			}
		},
	}
	listImages.Flags().Bool("tags", false, "List also tags for each image")
	listImages.Flags().IntP("concurrency", "x", 8, "Number of goroutines to use")
	list.AddCommand(listImages)
	list.AddCommand(listProjects)
}
