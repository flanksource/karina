package cmd

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

var Deletion = &cobra.Command{
	Use:   "delete",
	Short: "delete running pods",
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatal(err)
		}

		pods, _ := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		for _, pod := range pods.Items {
			// check if pod is running
			if *pod.Status.ContainerStatuses[0].Started {
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

