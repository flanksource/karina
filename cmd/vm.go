package cmd

import (
	"fmt"
	"strings"

	"github.com/flanksource/karina/pkg/provision"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const tagAnnotation = "tags.vsphere.flanksource.com"

func init() {
	tag := &cobra.Command{
		Use:   "tag",
		Short: "Tag virtual machines",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			cluster, err := provision.GetCluster(platform)
			if err != nil {
				log.Fatalf("failed to get cluster: %v", err)
			}

			for _, nm := range cluster.Nodes {
				tags := map[string]string{}

				for k, v := range nm.Node.Annotations {
					if strings.HasPrefix(k, tagAnnotation) {
						categoryID := strings.ReplaceAll(k, tagAnnotation+"/", "")
						tags[categoryID] = v
					}
				}

				if len(tags) == 0 {
					continue
				}

				vm := nm.Machine
				message := fmt.Sprintf("Name: %s Tags: ", vm.Name())
				for k, v := range tags {
					message += fmt.Sprintf("%s=%s ", k, v)
				}
				platform.Infof(message)

				if err := vm.SetTags(tags); err != nil {
					platform.Errorf("Failed to set tags for node %s: %v", nm.Node.Name, err)
				}
			}
		},
	}

	tag.Flags().String("vm", "", "The name of the VM")
	MachineImages.AddCommand(tag)
}
