package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CleanupJobs removes all failed jobs in a given namespace
var CleanupJobs = &cobra.Command{
	Use:   "cleanupjobs",
	Short: "remove all failed jobs in a given namespace",
	Run: func(cmd *cobra.Command, args []string) {

		namespace, _ := cmd.Flags().GetString("namespace")

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
						if err = clientSet.BatchV1().Jobs(namespace).Delete(context.TODO(), j.Name, metav1.DeleteOptions{}); err != nil {
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

func init() {
	CleanupJobs.Flags().String("namespace", "", "Namespace to cleanup failed jobs.")
}
