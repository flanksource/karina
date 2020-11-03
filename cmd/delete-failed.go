package cmd

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

var Deletion = &cobra.Command{
	Use:   "delete",
	Short: "delete running pods",
	Run: func(cmd *cobra.Command, args []string) {
		p := getPlatform(cmd)

		clientset, err :=  p.GetClientset()
		if err != nil {
			log.Fatal(err)
		}
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, pod := range pods.Items {
			// length check in place because not pods have values in the containerstatuses array
			if  len(pod.Status.ContainerStatuses) >= 1 && pod.Status.ContainerStatuses[0].State.Running != nil {
				//individual options for delete
				deleteOptions := metav1.DeleteOptions{}
				// delete each pod
				err := clientset.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &deleteOptions)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

	},
}

func init() {
}

