package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/batch/v1"
	v1p "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var Cleanup = &cobra.Command{
	Use:   "cleanup",
	Short: "remove all failed jobs or pods",
}

func init() {
	var all bool
	// CleanupJobs removes all failed jobs in a given namespace
	Jobs := &cobra.Command{
		Use:   "jobs",
		Short: "remove all failed jobs in a given namespace",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, _ := cmd.Flags().GetString("namespace")

			if all {
				namespace = metav1.NamespaceAll
			}

			p := getPlatform(cmd)
			clientSet, err := p.GetClientset()
			if err != nil {
				p.Fatalf("Failed to create the new k8s client: %v", err)
			}

			// gather the list of jobs from a namespace
			jobs, err := clientSet.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				p.Fatalf("Failed to gather the list of jobs from namespace %v: %v", namespace, err)
			}

			// WaitGroup to synchronize go routines execution
			wg := sync.WaitGroup{}

			// loop through the list of jobs found
			for _, j := range jobs.Items {
				wg.Add(1)

				go func(j v1.Job, clientSet *kubernetes.Clientset) {
					// loop through the job object's status conditions
					for _, conditions := range j.Status.Conditions {
						// if the type of the JobCondition is equal to "Failed", delete the job
						if conditions.Type == "Failed" {
							p.Infof("Removing failed job %v from namespace %v. Failed reason: %v", j.Name, j.Namespace, conditions.Reason)
							if err = clientSet.BatchV1().Jobs(j.Namespace).Delete(context.TODO(), j.Name, metav1.DeleteOptions{}); err != nil {
								p.Errorf("Failed to delete job: %v", err)
							}
							break
						}
					}
					wg.Done()
				}(j, clientSet)
			}

			wg.Wait()
		},
	}
	Pods := &cobra.Command{
		Use:   "pods",
		Short: "Delete non running Pods",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			p := getPlatform(cmd)
			client, err := p.GetClientset()
			if err != nil {
				p.Fatalf("unable to get clientset: %v", err)
			}

			namespace, _ := cmd.Flags().GetString("namespace")
			if all {
				namespace = metav1.NamespaceAll
			}

			// gather the list of Pods from all
			pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: "status.phase!=Running"})
			if err != nil {
				p.Fatalf("Failed to gather the list of pods from namespace %v: %v", namespace, err)
			}

			// WaitGroup to synchronize go routines execution
			wg := sync.WaitGroup{}

			for _, po := range pods.Items {
				wg.Add(1)

				go func(po v1p.Pod, clientSet *kubernetes.Clientset) {
					p.Infof("Removing failed pod %v from namespace %v. Failed reason: %v", po.Name, po.Namespace, po.Status.Conditions[len(po.Status.Conditions)-1].Message)
					if err = clientSet.CoreV1().Pods(po.Namespace).Delete(context.TODO(), po.Name, metav1.DeleteOptions{}); err != nil {
						p.Errorf("Failed to delete pod: %v", err)
					}

					wg.Done()
				}(po, client)
			}

			wg.Wait()
		},
	}

	Jobs.Flags().StringP("namespace", "n", "", "Namespace to cleanup failed jobs.")
	Jobs.Flags().BoolVarP(&all, "all", "a", false, "cleanup failed jobs from all namespaces")
	Pods.Flags().StringP("namespace", "n", "", "Namespace to cleanup non running pods.")
	Pods.Flags().BoolVarP(&all, "all", "a", false, "cleanup non running pods from all namespaces")
	Cleanup.AddCommand(Jobs, Pods)
}
