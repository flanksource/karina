package constants

import (
	"github.com/flanksource/kommons"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

var RuntimeObjects = map[string]kommons.RuntimeObjectWithMetadata{
	"configmap":              &v1.ConfigMap{},
	"daemonset":              &extensionsv1beta1.DaemonSet{},
	"deployment":             &extensionsv1beta1.Deployment{},
	"namespace":              &v1.Namespace{},
	"node":                   &v1.Node{},
	"persistentvolumeclaims": &v1.PersistentVolumeClaim{},
	"persistentvolume":       &v1.PersistentVolume{},
	"pod":                    &v1.Pod{},
	"secret":                 &v1.Secret{},
	"service":                &v1.Service{},
	"statefulset":            &appsv1beta1.StatefulSet{},
}
