package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"gopkg.in/flanksource/yaml.v3"

	"github.com/flanksource/karina/pkg/phases"
	konfigadm "github.com/flanksource/konfigadm/pkg/types"
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
				var annotations = platform.Master.Annotations

				if nodePoolName, ok := node.Labels[constants.NodePoolLabel]; ok {
					if pool, ok := platform.Nodes[nodePoolName]; ok {
						annotations = pool.Annotations
					}
				}

				for k, v := range annotations {
					node.Annotations[k] = v
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

	var master bool
	var install bool
	var cloudInit bool
	var nodeGroup string
	var tokenExpiry time.Duration
	generateJoinCommand := &cobra.Command{
		Use:   "generate-join-commands",
		Short: "Generate a new node token",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			if platform.JoinEndpoint == "" {
				cfg, _ := platform.GetRESTConfig()
				platform.JoinEndpoint = strings.Replace(cfg.Host, "https://", "", 1)
			}
			var config *konfigadm.Config
			var err error
			if master {
				config, err = phases.CreateSecondaryMaster(platform)
			} else {
				config, err = phases.CreateWorker(nodeGroup, platform)
			}
			if err != nil {
				logger.Fatalf("Cannot generate worker config: %v", err)
			}

			if install {
				data, _ := yaml.Marshal(konfigadm.Config{
					Kubernetes: &konfigadm.KubernetesSpec{
						Version: platform.Kubernetes.Version,
					},
					ContainerRuntime: konfigadm.ContainerRuntime{
						Type: platform.Kubernetes.ContainerRuntime,
					},
				})
				config.Files["/etc/kubernetes/konfigadm.yaml"] = string(data)
			}
			config.PreCommands = append(config.PreCommands, konfigadm.Command{
				Cmd: "wget -O konfigadm https://github.com/flanksource/konfigadm/releases/latest/download/konfigadm && chmod +x konfigadm && mv konfigadm /usr/bin",
			}, konfigadm.Command{
				Cmd: "konfigadm apply -c /etc/kubernetes/konfigadm.yaml",
			})

			fmt.Println(config.ToCloudInit().String())
		},
	}

	generateToken := &cobra.Command{
		Use:   "generate-token",
		Short: "Generate a new node token",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)
			token, err := kubeadm.GetOrCreateBootstrapToken(platform, tokenExpiry)
			if err != nil {
				logger.Fatalf("Error creating join token: %v", err)
			}
			fmt.Println(token)
		},
	}

	generateJoinCommand.Flags().BoolVar(&master, "master", false, "Create a new secondary master")
	generateJoinCommand.Flags().BoolVar(&install, "install", false, "Install kubernetes dependencies on boot")
	generateJoinCommand.Flags().StringVar(&nodeGroup, "node-group", "default", "")
	generateToken.Flags().DurationVar(&tokenExpiry, "token-expiry", 24*time.Hour, "")
	generateJoinCommand.Flags().BoolVar(&cloudInit, "cloud-init", true, "Generate cloud-init ")
	Node.AddCommand(annotate, ips, generateJoinCommand, generateToken)
}
