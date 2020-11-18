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
				tagsCh, err := harbor.ListImagesWithTags(platform, concurrency)
				if err != nil {
					log.Fatalf("Error listing images: %v", err)
				}

				for {
					tag, more := <-tagsCh
					if more {
						fmt.Printf("%s/%s:%s\n", tag.ProjectName, tag.RepositoryName, tag.Name)
					} else {
						break
					}
				}
			} else {
				imagesCh, err := harbor.ListImages(platform, concurrency)
				if err != nil {
					log.Fatalf("Error listing images: %v", err)
				}

				for {
					image, more := <-imagesCh
					if more {
						fmt.Printf("%s/%s\n", image.ProjectName, image.Name)
					} else {
						break
					}
				}
			}
		},
	}
	listImages.Flags().Bool("tags", false, "List also tags for each image")
	listImages.Flags().IntP("concurrency", "x", 8, "Number of goroutines to use")
	list.AddCommand(listImages)
	list.AddCommand(listProjects)

	integrityCheck := &cobra.Command{
		Use:   "integrity-check",
		Short: "Commands to check integrity of each manifest",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			concurrency, _ := cmd.Flags().GetInt("concurrency")

			brokenTagsCh, err := harbor.IntegrityCheck(platform, concurrency)

			if err != nil {
				log.Fatalf("Error checking tags: %v", err)
			}

			for {
				tag, more := <-brokenTagsCh
				if more {
					fmt.Printf("broken: %s/%s:%s\n", tag.ProjectName, tag.RepositoryName, tag.Name)
				} else {
					break
				}
			}
		},
	}
	integrityCheck.Flags().IntP("concurrency", "x", 8, "Number of goroutines to use")
	Harbor.AddCommand(integrityCheck)

	integrityCheckFromFile := &cobra.Command{
		Use:   "integrity-check-from-file",
		Short: "Commands to check integrity of each manifest",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			concurrency, _ := cmd.Flags().GetInt("concurrency")
			filename, _ := cmd.Flags().GetString("filename")

			brokenTagsCh, err := harbor.IntegrityCheckFromFile(platform, concurrency, filename)

			if err != nil {
				log.Fatalf("Error checking tags: %v", err)
			}

			for {
				tag, more := <-brokenTagsCh
				if more {
					fmt.Printf("broken: %s/%s:%s\n", tag.ProjectName, tag.RepositoryName, tag.Name)
				} else {
					break
				}
			}
		},
	}
	integrityCheckFromFile.Flags().IntP("concurrency", "x", 8, "Number of goroutines to use")
	integrityCheckFromFile.Flags().StringP("filename", "f", "", "Filename to parse broken tag from")
	Harbor.AddCommand(integrityCheckFromFile)
}
