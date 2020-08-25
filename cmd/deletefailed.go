package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DeleteFailed = &cobra.Command{
	Use:   "deletefailed",
	Short: "Deletes the pods which are not in Running state",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ns, _ := cmd.Flags().GetString("namespace")
		p := getPlatform(cmd)
		client, err := p.GetClientset()
		if err != nil {
			log.Fatal(err)
		}
		pod, err := client.CoreV1().Pods(ns).List(v1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		background := v1.DeletePropagationBackground
		for _, p := range pod.Items {
			for _, condition := range p.Status.Conditions {
				if condition.Type == "Ready" {
					// Crashoff on containers shows the Status.Phase as Running for pod, to avoid that condition.Status is used in conjugation.
					if p.Status.Phase != "Running" || condition.Status == "False" {
						err = client.CoreV1().Pods(p.Namespace).Delete(p.GetName(), &v1.DeleteOptions{PropagationPolicy: &background})
						if err != nil {
							fmt.Println("Error while Deleting POD", err)
						} else {
							fmt.Println(p.Name, "Deleted")
						}
					}
				}
			}
		}

	},
}

func init() {
	DeleteFailed.Flags().StringP("namespace", "n", "", "Name name. If omitted, command will be run across all pods matching selector")
}
