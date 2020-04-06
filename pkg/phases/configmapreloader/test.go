package configmapreloader

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
	if p.ConfigMapReloader.Disabled {
		test.Skipf("configmap-reloader", "configmap-reloader not configured")
		return
	}
	k8s.TestNamespace(client, Namespace, test)
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
	if p.ConfigMapReloader.Disabled {
		test.Skipf("configmap-reloader", "configmap-reloader not configured")
		return
	}
	defer cleanup(client)
	_, err := client.CoreV1().ConfigMaps(Namespace).Create(&v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reloader-test",
			Namespace: Namespace,
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
		test.Failf("configmap-reloader", "Cannot create configmap-reload config map")
		return
	}
	var replicas int32 = 1
	_, err = client.AppsV1().Deployments(Namespace).Create(&appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "configmap-reloader-test",
			Namespace: Namespace,
			Labels: map[string]string{
				"k8s-app": "configmap-reloader-test",
			},
			Annotations: map[string]string{
				"reload/configmap": "reloader-test",
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
		test.Failf("configmap-reloader", "Cannot create test deployment")
		return
	}

	watch, _ := client.AppsV1().Deployments(Namespace).Watch(metav1.ListOptions{
		LabelSelector:  "k8s-app=configmap-reloader-test",
		TimeoutSeconds: &watchTimeout,
	})
	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			log.Errorf("unexpected type")
		}
		if p.Status.ReadyReplicas == 1 {
			log.Tracef("Deployment is ready")
			break
		}
	}

	_, err = client.CoreV1().ConfigMaps(Namespace).Update(&v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reloader-test",
			Namespace: Namespace,
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
		test.Failf("configmap-reloader", "ConfigMap configmap-reloader was not updated: %v", err)
		return
	}
	log.Tracef("Updated ConfigMap")

	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			log.Fatal("unexpected type")
		}
		for _, condition := range p.Status.Conditions {
			if condition.Reason == "ReplicaSetUpdated" {
				test.Passf("configmap-reloader", "[e2e] configmap-reloader: new secret is available in recreated pods")
				return
			}
		}
	}
	test.Failf("configmap-reloader", "Deployment was not updated for %d seconds", watchTimeout)
}

//nolint
func cleanup(client *kubernetes.Clientset) {
	client.CoreV1().ConfigMaps(Namespace).Delete("reloader-test", &metav1.DeleteOptions{})
	client.AppsV1().Deployments(Namespace).Delete("configmap-reloader-test", &metav1.DeleteOptions{})

}
