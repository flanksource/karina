package cmd

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Exec = &cobra.Command{
	Use:   "exec",
	Short: "Execute a shell command inside pods matching selector",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		command := args[0]
		args = args[1:]
		if len(args) > 0 {
			command = command + " " + strings.Join(args, " ")
		}
		ns, _ := cmd.Flags().GetString("namespace")
		container, _ := cmd.Flags().GetString("container")
		selector, _ := cmd.Flags().GetString("selector")
		p := getPlatform(cmd)
		client, err := p.GetClientset()
		if err != nil {
			log.Fatalf("unable to get clientset: %v", err)
		}

		pods, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			log.Fatalf("unable to list pods: %v", err)
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != v1.PodRunning {
				log.Warnf("Skipping %s in %s phase", pod.Name, pod.Status.Phase)
				continue
			}
			_container := container
			if _container == "" {
				_container = pod.Spec.Containers[0].Name
			}
			stdout, stderr, err := p.ExecutePodf(ns, pod.Name, _container, command)
			if err != nil {
				log.Errorf("[%s/%s] %s %s %v", pod.Name, _container, stdout, stderr, err)
			} else {
				log.Infof("[%s/%s] %s %s", pod.Name, _container, stdout, stderr)
			}
		}
	},
}

var ExecNode = &cobra.Command{
	Use:   "exec-node",
	Short: "Execute a shell command inside host mounted daemonset on each node",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		args = args[1:]
		if len(args) > 0 {
			command = command + " " + strings.Join(args, " ")
		}
		selector, _ := cmd.Flags().GetString("selector")
		p := getPlatform(cmd)
		client, err := p.GetClientset()
		if err != nil {
			log.Fatalf("unable to get clientset: %v", err)
		}

		nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			log.Fatalf("unable to list nodes: %v", err)
		}

		for _, node := range nodes.Items {
			stdout, err := p.Executef(node.Name, 60*time.Second, command)
			if err != nil {
				log.Errorf("[%s] %s %v", node.Name, stdout, err)
			} else {
				log.Infof("[%s] %s", node.Name, stdout)
			}
		}
	},
}

func init() {
	Exec.Flags().StringP("namespace", "n", "", "Name name. If omitted, command will be run across all pods matching selector")
	Exec.Flags().StringP("container", "", "", "Container name. If omitted, the first container in the pod will be chosen")
	Exec.Flags().StringP("selector", "l", "", "Pod selector")
	ExecNode.Flags().StringP("selector", "l", "", "node selector")
}
