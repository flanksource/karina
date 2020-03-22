package configmapReloader

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var watchTimeout = int64(30) // Wait for deployment to update only N seconds

func Test(p *platform.Platform, test *console.TestResults, args []string, cmd *cobra.Command) {
	client, _ := p.GetClientset()
	if p.ConfigMapReloader == nil || p.ConfigMapReloader.Disabled {
		test.Skipf("configmap-reloader", "configmap-reloader not configured")
		return
	}
	k8s.TestNamespace(client, "configmap-reloader", test)
	runE2E, err := cmd.Flags().GetBool("e2e")
	if err != nil {
		return
	}
	if runE2E {
		e2eTest(p, test)
	}
}

func e2eTest(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.ConfigMapReloader == nil || p.ConfigMapReloader.Disabled {
		test.Skipf("configmap-reloader", "configmap-reloader not configured")
		return
	}
	_, err := client.CoreV1().ConfigMaps("configmap-reloader").Create(&v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reloader-test",
			Namespace: "configmap-reloader",
			Labels: map[string]string{
				"k8s-app":  "configmap-reloader",
				"e2e-test": "true",
			},
		},
		Data: map[string]string{
			"test": "Before reload",
		},
	})
	if err != nil {
		test.Failf("TestConfigMapFailed", "Cannot create configmap-reload config map")
		cleanup(client) // nolint: errcheck
	}
	var replicas int32 = 1
	_, err = client.AppsV1().Deployments("configmap-reloader").Create(&appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "configmap-reloader-test",
			Namespace: "configmap-reloader",
			Labels: map[string]string{
				"k8s-app": "configmap-reloader-test",
			},
			Annotations: map[string]string{
				"configmap.reloader.stakater.com/reload": "reloader-test",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "configmap-reloader-test",
				},
			},

			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "configmap-reloader-test",
					},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{{
						Name: "test-configmap",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "reloader-test",
								},
							},
						},
					}},
					Containers: []v1.Container{{
						Name:  "configmap-reloader-test",
						Image: "docker.io/nginx",
						VolumeMounts: []v1.VolumeMount{{
							Name:      "test-configmap",
							ReadOnly:  false,
							MountPath: "/var/lib/",
						},
						}},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	})

	if err != nil {
		test.Failf("TestDeploymentFailed", "Cannot create test deployment")
		cleanup(client) // nolint: errcheck
	}
	test.Passf("TestDeploymentCreated", "Created test deployment")

	watch, _ := client.AppsV1().Deployments("configmap-reloader").Watch(metav1.ListOptions{
		LabelSelector:  "k8s-app=configmap-reloader-test",
		TimeoutSeconds: &watchTimeout,
	})
	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			log.Fatal("unexpected type")
		}
		if p.Status.ReadyReplicas == 1 {
			log.Infof("Deployment is ready")
			break
		}
	}

	_, err = client.CoreV1().ConfigMaps("configmap-reloader").Update(&v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reloader-test",
			Namespace: "configmap-reloader",
			Labels: map[string]string{
				"k8s-app":  "configmap-reloader",
				"e2e-test": "true",
			},
		},
		Data: map[string]string{
			"test": "After reload",
		},
	})
	if err != nil {
		test.Failf("ConfigMapNotUpdated", "ConfigMap configmap-reloader was not updated: %v", err)
	}
	log.Infof("Updated ConfigMap")

	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			log.Fatal("unexpected type")
		}
		for _, condition := range p.Status.Conditions {
			if condition.Reason == "ReplicaSetUpdated" {
				test.Passf("DeploymentUpdated", "New secret is available in recreated pods")
				err := cleanup(client)
				if err != nil {
					log.Fatal("Failed to delete test resources")
				}
				return
			}
		}
	}
	test.Failf("DeploymentNotUpdated", "Deployment was not updated for %d seconds", watchTimeout)
	err = cleanup(client)
	if err != nil {
		log.Fatalf("Failed to delete test resources: %v", err)
	}
}

func cleanup(client *kubernetes.Clientset) error {
	err := client.CoreV1().ConfigMaps("configmap-reloader").Delete("reloader-test", &metav1.DeleteOptions{})
	log.Errorf("failed to delete test configmap: %v", err)
	err = client.AppsV1().Deployments("configmap-reloader").Delete("configmap-reloader-test", &metav1.DeleteOptions{})
	log.Errorf("failed to delete test deployment: %v", err)
	return nil
}
