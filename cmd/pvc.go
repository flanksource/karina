package cmd

import (
	"context"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/errors"
)

var Pvc = &cobra.Command{
	Use:   "pvc",
	Short: "Commands for interacting with Kubernetes persistent volume claims",
}

var pvcname, pvcsize, pvcnamespace, pvcclass string

func init() {
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new PVC of the given name and size",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			pvcname, _ = cmd.Flags().GetString("name")
			pvcsize, _ = cmd.Flags().GetString("size")
			pvcnamespace, _ = cmd.Flags().GetString("namespace")
			pvcclass, _ = cmd.Flags().GetString("class")

			if pvcname == "" {
				platform.Fatalf("Provide a name to create PVC!")
			}

			if pvcsize == "" {
				platform.Fatalf("Provide a size to create PVC!")
			}

			if pvcclass == "" {
				platform.Fatalf("Provide a storage class to create PVC!")
			}

			err := platform.GetOrCreatePVC(pvcnamespace, pvcname, pvcsize, pvcclass)
			if err != nil {
				platform.Fatalf("Failed to create PVC (%s): %s", pvcname, err)
			}
		},
	}

	create.Flags().String("name", "", "The name of the PVC")
	create.Flags().String("size", "", "The size of the PVC")
	create.Flags().String("namespace", "default", "The namespace of the PVC (default: default)")
	create.Flags().String("class", "", "The storage class of the PVC (default: none)")

	resize := &cobra.Command{
		Use:   "resize",
		Short: "Resize a PVC",
		Run: func(cmd *cobra.Command, args []string) {
			platform := getPlatform(cmd)

			pvcname, _ = cmd.Flags().GetString("name")
			pvcsize, _ = cmd.Flags().GetString("size")
			pvcnamespace, _ = cmd.Flags().GetString("namespace")
			pvcclass, _ = cmd.Flags().GetString("class")

			if pvcname == "" {
				platform.Fatalf("Provide a name to resize PVC!")
			}

			if pvcsize == "" {
				platform.Fatalf("Provide a size to resize PVC!")
			}

			clientset, err := platform.GetClientset()
			if err != nil {
				platform.Fatalf("failed to get clientset: %s", err)
			}

			qty, err := resource.ParseQuantity(pvcsize)
			if err != nil {
				platform.Errorf("failed to parse quantity: %v", err)
			}
			pvcs := clientset.CoreV1().PersistentVolumeClaims(pvcnamespace)
		
			existing, err := pvcs.Get(context.TODO(), pvcname, metav1.GetOptions{})
			if err != nil && errors.IsNotFound(err) {
				platform.Fatalf("failed to get PVC: %s", err)
			} else {
				//TODO: Fetch PVC details by name and make sure we resize the correct one.
				platform.Infof("Found existing PVC %s/%s ==> %s\n", pvcnamespace, pvcname, existing.UID)
				platform.Infof("Resizing PVC %s/%s (%s %s)\n", pvcnamespace, pvcname, pvcsize, pvcclass)
				_, err = pvcs.Update(context.TODO(), &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: pvcname,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						// StorageClassName: &pvcclass,
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: qty,
							},
						},
					},
				}, metav1.UpdateOptions{})
			}
		
			if err != nil {
				platform.Fatalf("Failed to resize PVC (%s): %s", pvcname, err)
			}
		},
	}

	resize.Flags().String("name", "", "The name of the PVC")
	resize.Flags().String("size", "", "The size of the PVC")
	resize.Flags().String("namespace", "default", "The namespace of the PVC (default: default)")
	resize.Flags().String("class", "", "The storage class of the PVC (default: none)")

	Pvc.AddCommand(create, resize)
}
