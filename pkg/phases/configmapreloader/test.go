package configmapreloader

import (
	"context"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var testNamespace = "default"
var testName = "configmap-reloader-test"
var watchTimeout = int64(60) // Wait for deployment to update only N seconds

func Test(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	if p.ConfigMapReloader.Disabled {
		return
	}
	if err := p.WaitForDeployment(constants.PlatformSystem, "reloader", 30*time.Second); err != nil {
		test.Failf(testName, "configmap-reloader did not come up")
		return
	}
	kommons.TestDeploy(client, constants.PlatformSystem, "reloader", test)
	if !p.E2E {
		return
	}
	e2eTest(p, test)
}

func e2eTest(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	//cleanup from any failed run
	cleanup(client)
	//cleanup correctly after the end of the run
	defer cleanup(client)

	if err := p.CreateOrUpdateConfigMap(testName, testNamespace, map[string]string{
		"test": "Before reload",
	}); err != nil {
		test.Failf(testName, "Cannot create configmap-reload config map: %v", err)
		return
	}

	var replicas int32 = 1

	_, err := client.AppsV1().Deployments("default").Create(context.TODO(), &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: "default",
			Labels: map[string]string{
				"k8s-app": testName,
			},
			Annotations: map[string]string{
				"reload/all": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": testName,
				},
			},

			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": testName,
					},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{{
						Name: "test-configmap",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: testName,
								},
							},
						},
					}},
					Containers: []v1.Container{{
						Name:  testName,
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
	}, metav1.CreateOptions{})

	if err != nil {
		test.Failf("configmap-reloader", "Cannot create test deployment: %v", err)
		return
	}

	watch, _ := client.AppsV1().Deployments(testNamespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector:  "k8s-app=configmap-reloader-test",
		TimeoutSeconds: &watchTimeout,
	})
	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			test.Errorf("unexpected type")
		}
		if p.Status.ReadyReplicas == 1 {
			break
		}
	}

	if err := p.CreateOrUpdateConfigMap(testName, testNamespace, map[string]string{
		"test": "After reload",
	}); err != nil {
		test.Failf(testName, "Cannot update configmap-reload config map: %v", err)
		return
	}

	for event := range watch.ResultChan() {
		p, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			test.Failf(testName, "unexpected type")
			return
		}
		for _, condition := range p.Status.Conditions {
			if condition.Reason == "ReplicaSetUpdated" {
				test.Passf(testName, "[e2e] configmap-reloader: new secret is available in recreated pods")
				return
			}
		}
	}
	test.Failf(testName, "Deployment was not updated for %d seconds", watchTimeout)
}

//nolint
func cleanup(client *kubernetes.Clientset) {
	client.CoreV1().ConfigMaps(testNamespace).Delete(context.TODO(), testName, metav1.DeleteOptions{})
	client.AppsV1().Deployments(testNamespace).Delete(context.TODO(), testName, metav1.DeleteOptions{})
}
