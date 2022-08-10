package quack

import (
	"context"
	"fmt"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	if platform.Quack.IsDisabled() {
		return
	}

	client, _ := platform.GetClientset()
	kommons.TestNamespace(client, Namespace, test)

	if !platform.E2E {
		return
	}

	namespace := fmt.Sprintf("quack-test-%s", utils.RandomKey(6))

	if err := platform.CreateOrUpdateWorkloadNamespace(namespace, EnabledLabels, nil); err != nil {
		test.Failf("quack", "Failed to create test namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	ingresses := client.NetworkingV1beta1().Ingresses(namespace)

	ingress, err := ingresses.Create(context.TODO(), &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{

			Rules: []v1beta1.IngressRule{
				{
					Host: namespace + ".{{.domain}}",
				},
			},
		},
	}, metav1.CreateOptions{})

	if err != nil {
		test.Failf("quack", "failed to create ingress: %v", err)
		return
	}

	exptecedHost := namespace + "." + platform.Domain
	if ingress.Spec.Rules[0].Host != exptecedHost {
		test.Failf("quack", "expected %s, got %s", exptecedHost, ingress.Spec.Rules[0].Host)
	} else {
		test.Passf("quack", "quack templated ingress successfully")
	}
}
