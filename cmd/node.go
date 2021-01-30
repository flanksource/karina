package cmd

import (
	"context"
	"fmt"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Node = &cobra.Command{
	Use:   "node",
	Short: "Commands for interacting with Kubernetes nodes",
}

func init() {
	annotate := &cobra.Command{
		Use:   "annotate",
		Short: "Annotate nodes using annotations on vm master / worker pools",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			clientset, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("failed to get clientset: %s", err)
			}

			nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				platform.Fatalf("failed to list nodes: %v", err)
			}

			for _, node := range nodes.Items {
				_, isMaster := node.Labels[constants.MasterNodeLabel]
				if isMaster {
					for k, v := range platform.Master.Annotations {
						node.Annotations[k] = v
					}
				} else {
					nodePool, found := node.Labels[constants.NodePoolLabel]
					if !found {
						continue
					}
					workerPool, found := platform.Nodes[nodePool]
					if !found {
						platform.Errorf("Node %s has node pool %s which was not found in platform.Nodes", node.Name, nodePool)
						continue
					}
					annotations := workerPool.Annotations
					for k, v := range annotations {
						node.Annotations[k] = v
					}
				}

				if _, err := clientset.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{}); err != nil {
					platform.Errorf("Failed to update node %s: %v", node, err)
					continue
				}

				platform.Infof("Node %s annotated", node.Name)
			}
		},
	}

	ips := &cobra.Command{
		Use:   "ips",
		Short: "List all internal node IP's",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			clientset, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("failed to get clientset: %s", err)
			}

			nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				platform.Fatalf("failed to list nodes: %v", err)
			}

			for _, node := range nodes.Items {

				for _, address := range node.Status.Addresses {
					if address.Type == "InternalIP" {
						fmt.Println(address.Address)
					}
				}
			}
		},
	}

	Node.AddCommand(annotate, ips)
}
