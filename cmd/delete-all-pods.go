package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The deletion of pods gives the potentiality to deploys pods to new nodes
var DeleteAllPods = &cobra.Command{
	Use:   "delete-all-pods",
	Short: "Delete all pods",
	Args:  cobra.MinimumNArgs(0),
  Run: func(cmd *cobra.Command, args []string) {

		p := getPlatform(cmd)
		client, err := p.GetClientset()
		if err != nil {
			log.Fatalf("unable to get clientset: %v", err)
		}

    // Get all pods information from all namespaces
    pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})

    if err != nil {
      panic(err.Error())
    }

    if len(pods.Items) > 0 {
      for _, pod := range pods.Items {
        // List Pods
        fmt.Printf("Deleting pod: %s\n", pod.GetName())
        // Delete Pods
        if err = client.CoreV1().Pods(pod.GetNamespace()).Delete(pod.GetName(), nil); err != nil {
          p.Errorf("failed to delete pod %s", pod.GetName())
        }
      }
    } else {
      fmt.Println("No pods found!")
    }
  },
}
