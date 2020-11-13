package cmd

import (
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

	listImages := &cobra.Command{
		Use:   "images",
		Short: "List images in harbor",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			project, _ := cmd.Flags().GetString("project")
			listTags, _ := cmd.Flags().GetBool("tags")

			if project == "" {
				log.Fatalf("Please provide harbor project")
			}

			if err := harbor.ListImages(platform, project, listTags); err != nil {
				log.Fatalf("Error listing images: %v", err)
			}
		},
	}
	listImages.Flags().String("project", "", "Harbor project name")
	listImages.Flags().Bool("tags", false, "List also tags for each image")
	list.AddCommand(listImages)
}
