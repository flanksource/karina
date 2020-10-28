package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
)
import "k8s.io/client-go/tools/clientcmd"

var Deletion = &cobra.Command{
	Use:   "delete",
	Short: "Commands for deletions",
}

func init() {
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
	var ns, label, field string
	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.StringVar(&label, "l", "", "Label selector")
	flag.StringVar(&field, "f", "", "Field selector")
	api := clientset.CoreV1()

	pods, _ := clientset.CoreV1().Pods("kubernetes").List(metav1.ListOptions{FieldSelector: "metadata.name=kubernetes"})
	for _, pod := range pods.Items {
		fmt.Println(pod.Name, pod.Status)
	}

	os.Exit(31)

	listOptions := metav1.ListOptions{
		LabelSelector: label,
		FieldSelector: field,
	}
	fmt.Printf("\n\n\n ns ::: %s, label :: %s , field ::: %s \n\n\n",ns, label, field )
	os.Exit(32)


	pvcs, err := api.PersistentVolumeClaims(ns).List(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	printPVCs(pvcs)


	os.Exit(33)
}


func printPVCs(pvcs *v1.PersistentVolumeClaimList) {
	template := "%-32s%-8s%-8s\n"
	fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
	for _, pvc := range pvcs.Items {
		quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		fmt.Printf(
			template,
			pvc.Name,
			string(pvc.Status.Phase),
			quant.String())
	}
}